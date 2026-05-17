package controllers

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"

	quote "gateway/quote"
	users "gateway/users"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	QuoteCodeNotFound        int32 = 1001
	QuoteCodeAlreadyExists   int32 = 1002
	QuoteCodeInvalidInput    int32 = 1003
	QuoteCodeInvalidLineType int32 = 1004
	QuoteCodeInvalidLineData int32 = 1005
	QuoteCodeFinalized       int32 = 1006
	QuoteCodeInternalError   int32 = 2001
)

var quoteErrors = &serviceErrors{
	codes: map[int32]codeMapping{
		QuoteCodeNotFound:        {http.StatusNotFound, "Devis introuvable."},
		QuoteCodeAlreadyExists:   {http.StatusConflict, "Cette ressource existe déjà."},
		QuoteCodeInvalidInput:    {http.StatusBadRequest, "Données invalides."},
		QuoteCodeInvalidLineType: {http.StatusBadRequest, "Type de ligne invalide."},
		QuoteCodeInvalidLineData: {http.StatusBadRequest, "Données de ligne invalides."},
		QuoteCodeFinalized:       {http.StatusConflict, "Ce devis est finalisé et ne peut plus être modifié."},
		QuoteCodeInternalError:   {http.StatusInternalServerError, "Une erreur interne est survenue."},
	},
	unavailableMessage: "Service devis indisponible.",
}

// QuotesRoutes wires the /quotes API group against the quote gRPC service.
func QuotesRoutes(r *gin.RouterGroup) {
	address := os.Getenv("QUOTE_SERVICE_ADDRESS")
	if address == "" {
		address = "localhost:50053"
	}
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to quote gRPC server: %v", err)
	}
	client := quote.NewQuoteServiceClient(conn)

	usersAddress := os.Getenv("USER_SERVICE_ADDRESS")
	if usersAddress == "" {
		usersAddress = "localhost:50052"
	}
	usersConn, err := grpc.NewClient(usersAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to users gRPC server: %v", err)
	}
	usersClient := users.NewUserServiceClient(usersConn)

	r.GET("", func(c *gin.Context) { ListQuotes(c, client, usersClient) })
	r.POST("", func(c *gin.Context) { CreateQuote(c, client) })

	archive := r.Group("/archive")
	archive.DELETE("/trash", func(c *gin.Context) { TrashQuotes(c, client) })

	one := r.Group("/:id")
	one.GET("", func(c *gin.Context) { GetQuote(c, client) })
	one.PUT("", func(c *gin.Context) { UpdateQuote(c, client) })
	one.DELETE("", func(c *gin.Context) { DeleteQuote(c, client) })
	one.POST("/archive", func(c *gin.Context) { ArchiveQuote(c, client) })
	one.POST("/restore", func(c *gin.Context) { RestoreQuote(c, client) })
	one.POST("/drop", func(c *gin.Context) { DropQuote(c, client) })
	one.POST("/continue", func(c *gin.Context) { ContinueQuote(c, client) })

	lines := one.Group("/lines")
	lines.GET("", func(c *gin.Context) { ListQuoteLines(c, client) })
	lines.POST("", func(c *gin.Context) { CreateQuoteLine(c, client) })
	lines.GET("/:lineId", func(c *gin.Context) { GetQuoteLine(c, client) })
	lines.PUT("/:lineId", func(c *gin.Context) { UpdateQuoteLine(c, client) })
	lines.DELETE("/:lineId", func(c *gin.Context) { DeleteQuoteLine(c, client) })
}

// ─── Quote handlers ──────────────────────────────────────────────────────────

func ListQuotes(c *gin.Context, client quote.QuoteServiceClient, usersClient users.UserServiceClient) {
	includeArchived := c.Query("archived") == "true"
	userID := userIDFromCtx(c)

	var (
		quotesResp *quote.ListQuotesResponse
		linesResp  *quote.ListUserQuoteLinesResponse
	)

	g, gctx := errgroup.WithContext(c.Request.Context())
	g.Go(func() error {
		resp, err := client.ListQuotes(gctx, &quote.ListQuotesRequest{
			UserId:          userID,
			IncludeArchived: includeArchived,
		})
		if err != nil {
			return err
		}
		quotesResp = resp
		return nil
	})
	g.Go(func() error {
		resp, err := client.ListUserQuoteLines(gctx, &quote.ListUserQuoteLinesRequest{
			UserId:          userID,
			IncludeArchived: includeArchived,
		})
		if err != nil {
			return err
		}
		linesResp = resp
		return nil
	})
	if err := g.Wait(); err != nil {
		quoteErrors.unavailable(c)
		return
	}
	if !quotesResp.Success {
		quoteErrors.reply(c, quotesResp.Code)
		return
	}

	// Fetch taxes after lines so we can pass the referenced tax_ids as
	// include_ids — otherwise superseded taxes are filtered out by
	// users.ListTaxesForUser and lines using them would silently fall back to
	// 0% (TTC == HT). See backend/users/actions/tax/list_for_user.go.
	taxesResp, err := usersClient.ListTaxesForUser(c.Request.Context(), &users.ListTaxesForUserRequest{
		UserId:     userID,
		IncludeIds: distinctTaxIds(linesResp.Lines),
	})
	if err != nil {
		quoteErrors.unavailable(c)
		return
	}

	totals := computeQuoteTotals(linesResp.Lines, taxesResp.Taxes)

	out := make([]gin.H, 0, len(quotesResp.Quotes))
	for _, q := range quotesResp.Quotes {
		m := marshalQuote(q)
		m["total_ttc"] = totals[q.QuoteId]
		out = append(out, m)
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "quotes": out})
}

func distinctTaxIds(lines []*quote.QuoteLine) []int32 {
	seen := make(map[int32]struct{}, len(lines))
	ids := make([]int32, 0, len(lines))
	for _, l := range lines {
		if l.TaxId == 0 {
			continue
		}
		if _, ok := seen[l.TaxId]; ok {
			continue
		}
		seen[l.TaxId] = struct{}{}
		ids = append(ids, l.TaxId)
	}
	return ids
}

// computeQuoteTotals folds quote_lines into per-quote TTC totals in cents.
// Round once per line — matches the per-line breakdown of quote-form.tsx and
// the canonical PDF total. Lines with tax_id=0 ("no tax") contribute HT only.
func computeQuoteTotals(lines []*quote.QuoteLine, taxes []*users.Tax) map[string]int64 {
	rates := make(map[int32]float64, len(taxes)+1)
	for _, t := range taxes {
		r, err := strconv.ParseFloat(t.Rate, 64)
		if err != nil {
			continue
		}
		rates[t.Id] = r
	}

	totals := map[string]int64{}
	for _, l := range lines {
		qty, err := strconv.ParseFloat(l.Quantity, 64)
		if err != nil {
			continue
		}
		rate := rates[l.TaxId]
		totals[l.QuoteId] += int64(math.Round(qty * float64(l.UnitPrice) * (1 + rate/100)))
	}
	return totals
}

func CreateQuote(c *gin.Context, client quote.QuoteServiceClient) {
	var input struct {
		Name          string `json:"name" binding:"required"`
		ClientID      string `json:"client_id" binding:"required"`
		AddressID     int32  `json:"address_id" binding:"required"`
		UserAddressID int32  `json:"user_address_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.CreateQuote(c.Request.Context(), &quote.CreateQuoteRequest{
		UserId:        userIDFromCtx(c),
		Name:          input.Name,
		ClientId:      input.ClientID,
		AddressId:     input.AddressID,
		UserAddressId: input.UserAddressID,
	})
	if err != nil {
		quoteErrors.unavailable(c)
		return
	}
	if !resp.Success {
		quoteErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "quote_id": resp.QuoteId})
}

func GetQuote(c *gin.Context, client quote.QuoteServiceClient) {
	resp, err := client.GetQuote(c.Request.Context(), &quote.GetQuoteRequest{
		QuoteId: c.Param("id"),
		UserId:  userIDFromCtx(c),
	})
	if err != nil {
		quoteErrors.unavailable(c)
		return
	}
	if !resp.Success {
		quoteErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"quote":   marshalQuote(resp.Quote),
		"lines":   marshalLines(resp.Lines),
	})
}

func UpdateQuote(c *gin.Context, client quote.QuoteServiceClient) {
	var input struct {
		Name          string `json:"name" binding:"required"`
		ClientID      string `json:"client_id"`
		AddressID     int32  `json:"address_id"`
		UserAddressID int32  `json:"user_address_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.UpdateQuote(c.Request.Context(), &quote.UpdateQuoteRequest{
		QuoteId:       c.Param("id"),
		UserId:        userIDFromCtx(c),
		Name:          input.Name,
		ClientId:      input.ClientID,
		AddressId:     input.AddressID,
		UserAddressId: input.UserAddressID,
	})
	if err != nil {
		quoteErrors.unavailable(c)
		return
	}
	if !resp.Success {
		quoteErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func DeleteQuote(c *gin.Context, client quote.QuoteServiceClient) {
	resp, err := client.DeleteQuote(c.Request.Context(), &quote.DeleteQuoteRequest{
		QuoteId: c.Param("id"),
		UserId:  userIDFromCtx(c),
	})
	if err != nil {
		quoteErrors.unavailable(c)
		return
	}
	if !resp.Success {
		quoteErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func ArchiveQuote(c *gin.Context, client quote.QuoteServiceClient) {
	resp, err := client.ArchiveQuote(c.Request.Context(), &quote.ArchiveQuoteRequest{
		QuoteId: c.Param("id"),
		UserId:  userIDFromCtx(c),
	})
	if err != nil {
		quoteErrors.unavailable(c)
		return
	}
	if !resp.Success {
		quoteErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func RestoreQuote(c *gin.Context, client quote.QuoteServiceClient) {
	resp, err := client.RestoreQuote(c.Request.Context(), &quote.RestoreQuoteRequest{
		QuoteId: c.Param("id"),
		UserId:  userIDFromCtx(c),
	})
	if err != nil {
		quoteErrors.unavailable(c)
		return
	}
	if !resp.Success {
		quoteErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func TrashQuotes(c *gin.Context, client quote.QuoteServiceClient) {
	resp, err := client.TrashQuotes(c.Request.Context(), &quote.TrashQuotesRequest{
		UserId: userIDFromCtx(c),
	})
	if err != nil {
		quoteErrors.unavailable(c)
		return
	}
	if !resp.Success {
		quoteErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func DropQuote(c *gin.Context, client quote.QuoteServiceClient) {
	resp, err := client.DropQuote(c.Request.Context(), &quote.DropQuoteRequest{
		QuoteId: c.Param("id"),
		UserId:  userIDFromCtx(c),
	})
	if err != nil {
		quoteErrors.unavailable(c)
		return
	}
	if !resp.Success {
		quoteErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func ContinueQuote(c *gin.Context, client quote.QuoteServiceClient) {
	resp, err := client.ContinueQuote(c.Request.Context(), &quote.ContinueQuoteRequest{
		QuoteId: c.Param("id"),
		UserId:  userIDFromCtx(c),
	})
	if err != nil {
		quoteErrors.unavailable(c)
		return
	}
	if !resp.Success {
		quoteErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ─── Line handlers ───────────────────────────────────────────────────────────

type lineInput struct {
	Type      string          `json:"type" binding:"required"`
	Name      string          `json:"name"`
	Quantity  string          `json:"quantity" binding:"required"`
	Unit      string          `json:"unit"`
	UnitPrice int64           `json:"unit_price"`
	Data      json.RawMessage `json:"data"`
	Position  int32           `json:"position"`
	TaxID     int32           `json:"tax_id"`
}

func (in lineInput) dataString() string {
	if len(in.Data) == 0 {
		return ""
	}
	return string(in.Data)
}

func ListQuoteLines(c *gin.Context, client quote.QuoteServiceClient) {
	resp, err := client.ListQuoteLines(c.Request.Context(), &quote.ListQuoteLinesRequest{
		QuoteId: c.Param("id"),
		UserId:  userIDFromCtx(c),
	})
	if err != nil {
		quoteErrors.unavailable(c)
		return
	}
	if !resp.Success {
		quoteErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "lines": marshalLines(resp.Lines)})
}

func CreateQuoteLine(c *gin.Context, client quote.QuoteServiceClient) {
	var input lineInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.CreateQuoteLine(c.Request.Context(), &quote.CreateQuoteLineRequest{
		QuoteId:   c.Param("id"),
		UserId:    userIDFromCtx(c),
		Type:      input.Type,
		Name:      input.Name,
		Quantity:  input.Quantity,
		Unit:      input.Unit,
		UnitPrice: input.UnitPrice,
		Data:      input.dataString(),
		Position:  input.Position,
		TaxId:     input.TaxID,
	})
	if err != nil {
		quoteErrors.unavailable(c)
		return
	}
	if !resp.Success {
		quoteErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "line_id": resp.LineId})
}

func GetQuoteLine(c *gin.Context, client quote.QuoteServiceClient) {
	resp, err := client.GetQuoteLine(c.Request.Context(), &quote.GetQuoteLineRequest{
		LineId: c.Param("lineId"),
		UserId: userIDFromCtx(c),
	})
	if err != nil {
		quoteErrors.unavailable(c)
		return
	}
	if !resp.Success {
		quoteErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "line": marshalLine(resp.Line)})
}

func UpdateQuoteLine(c *gin.Context, client quote.QuoteServiceClient) {
	var input lineInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.UpdateQuoteLine(c.Request.Context(), &quote.UpdateQuoteLineRequest{
		LineId:    c.Param("lineId"),
		UserId:    userIDFromCtx(c),
		Type:      input.Type,
		Name:      input.Name,
		Quantity:  input.Quantity,
		Unit:      input.Unit,
		UnitPrice: input.UnitPrice,
		Data:      input.dataString(),
		Position:  input.Position,
		TaxId:     input.TaxID,
	})
	if err != nil {
		quoteErrors.unavailable(c)
		return
	}
	if !resp.Success {
		quoteErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func DeleteQuoteLine(c *gin.Context, client quote.QuoteServiceClient) {
	resp, err := client.DeleteQuoteLine(c.Request.Context(), &quote.DeleteQuoteLineRequest{
		LineId: c.Param("lineId"),
		UserId: userIDFromCtx(c),
	})
	if err != nil {
		quoteErrors.unavailable(c)
		return
	}
	if !resp.Success {
		quoteErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// marshalQuote emits state as a lowercase string instead of the proto enum's
// integer value, so the frontend can use literal "draft"|"sent"|"validated"|"drop".
func marshalQuote(q *quote.Quote) gin.H {
	if q == nil {
		return nil
	}
	return gin.H{
		"quote_id":        q.QuoteId,
		"user_id":         q.UserId,
		"name":            q.Name,
		"archived":        q.Archived,
		"state":           stateToLower(q.State),
		"client_id":       q.ClientId,
		"address_id":      q.AddressId,
		"user_address_id": q.UserAddressId,
	}
}


func stateToLower(s quote.QuoteState) string {
	switch s {
	case quote.QuoteState_QUOTE_STATE_DRAFT:
		return "draft"
	case quote.QuoteState_QUOTE_STATE_SENT:
		return "sent"
	case quote.QuoteState_QUOTE_STATE_VALIDATED:
		return "validated"
	case quote.QuoteState_QUOTE_STATE_DROP:
		return "drop"
	default:
		return "draft"
	}
}

// nullableInt converts the proto3 sentinel (0 = unset) into a JSON-friendly
// pointer so wire-level "no value" stays distinct from "the literal zero".
func nullableInt(v int32) *int32 {
	if v == 0 {
		return nil
	}
	return &v
}

// marshalLine emits the raw JSON `data` field as an object instead of a string,
// so consumers don't have to double-decode.
func marshalLine(l *quote.QuoteLine) gin.H {
	if l == nil {
		return nil
	}
	out := gin.H{
		"line_id":    l.LineId,
		"quote_id":   l.QuoteId,
		"type":       l.Type,
		"name":       l.Name,
		"quantity":   l.Quantity,
		"unit":       l.Unit,
		"unit_price": l.UnitPrice,
		"position":   l.Position,
	}
	out["tax_id"] = nullableInt(l.TaxId)
	if l.Data == "" {
		out["data"] = json.RawMessage("{}")
	} else {
		out["data"] = json.RawMessage(l.Data)
	}
	return out
}

func marshalLines(lines []*quote.QuoteLine) []gin.H {
	out := make([]gin.H, 0, len(lines))
	for _, l := range lines {
		out = append(out, marshalLine(l))
	}
	return out
}
