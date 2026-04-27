package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	quote "gateway/quote"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	QuoteCodeNotFound        int32 = 1001
	QuoteCodeAlreadyExists   int32 = 1002
	QuoteCodeInvalidInput    int32 = 1003
	QuoteCodeInvalidLineType int32 = 1004
	QuoteCodeInvalidLineData int32 = 1005
	QuoteCodeInternalError   int32 = 2001
)

var quoteErrorMap = map[int32]struct {
	Status  int
	Message string
}{
	QuoteCodeNotFound:        {http.StatusNotFound, "Devis introuvable."},
	QuoteCodeAlreadyExists:   {http.StatusConflict, "Cette ressource existe déjà."},
	QuoteCodeInvalidInput:    {http.StatusBadRequest, "Données invalides."},
	QuoteCodeInvalidLineType: {http.StatusBadRequest, "Type de ligne invalide."},
	QuoteCodeInvalidLineData: {http.StatusBadRequest, "Données de ligne invalides."},
	QuoteCodeInternalError:   {http.StatusInternalServerError, "Une erreur interne est survenue."},
}

func quoteError(c *gin.Context, code int32) {
	if mapped, ok := quoteErrorMap[code]; ok {
		c.JSON(mapped.Status, gin.H{"success": false, "message": mapped.Message, "code": code})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Une erreur inconnue est survenue.", "code": code})
	}
}

func quoteUnavailable(c *gin.Context) {
	c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": "Service devis indisponible."})
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

	r.GET("", func(c *gin.Context) { ListQuotes(c, client) })
	r.POST("", func(c *gin.Context) { CreateQuote(c, client) })

	archive := r.Group("/archive")
	archive.DELETE("/trash", func(c *gin.Context) { TrashQuotes(c, client) })

	one := r.Group("/:id")
	one.GET("", func(c *gin.Context) { GetQuote(c, client) })
	one.PUT("", func(c *gin.Context) { UpdateQuote(c, client) })
	one.DELETE("", func(c *gin.Context) { DeleteQuote(c, client) })
	one.POST("/archive", func(c *gin.Context) { ArchiveQuote(c, client) })
	one.POST("/restore", func(c *gin.Context) { RestoreQuote(c, client) })

	lines := one.Group("/lines")
	lines.GET("", func(c *gin.Context) { ListQuoteLines(c, client) })
	lines.POST("", func(c *gin.Context) { CreateQuoteLine(c, client) })
	lines.GET("/:lineId", func(c *gin.Context) { GetQuoteLine(c, client) })
	lines.PUT("/:lineId", func(c *gin.Context) { UpdateQuoteLine(c, client) })
	lines.DELETE("/:lineId", func(c *gin.Context) { DeleteQuoteLine(c, client) })
}

// ─── Quote handlers ──────────────────────────────────────────────────────────

func ListQuotes(c *gin.Context, client quote.QuoteServiceClient) {
	includeArchived := c.Query("archived") == "true"
	resp, err := client.ListQuotes(c.Request.Context(), &quote.ListQuotesRequest{
		UserId:          userIDFromCtx(c),
		IncludeArchived: includeArchived,
	})
	if err != nil {
		quoteUnavailable(c)
		return
	}
	if !resp.Success {
		quoteError(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "quotes": resp.Quotes})
}

func CreateQuote(c *gin.Context, client quote.QuoteServiceClient) {
	var input struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.CreateQuote(c.Request.Context(), &quote.CreateQuoteRequest{
		UserId: userIDFromCtx(c),
		Name:   input.Name,
	})
	if err != nil {
		quoteUnavailable(c)
		return
	}
	if !resp.Success {
		quoteError(c, resp.Code)
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
		quoteUnavailable(c)
		return
	}
	if !resp.Success {
		quoteError(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"quote":   resp.Quote,
		"lines":   marshalLines(resp.Lines),
	})
}

func UpdateQuote(c *gin.Context, client quote.QuoteServiceClient) {
	var input struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.UpdateQuote(c.Request.Context(), &quote.UpdateQuoteRequest{
		QuoteId: c.Param("id"),
		UserId:  userIDFromCtx(c),
		Name:    input.Name,
	})
	if err != nil {
		quoteUnavailable(c)
		return
	}
	if !resp.Success {
		quoteError(c, resp.Code)
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
		quoteUnavailable(c)
		return
	}
	if !resp.Success {
		quoteError(c, resp.Code)
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
		quoteUnavailable(c)
		return
	}
	if !resp.Success {
		quoteError(c, resp.Code)
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
		quoteUnavailable(c)
		return
	}
	if !resp.Success {
		quoteError(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func TrashQuotes(c *gin.Context, client quote.QuoteServiceClient) {
	resp, err := client.TrashQuotes(c.Request.Context(), &quote.TrashQuotesRequest{
		UserId: userIDFromCtx(c),
	})
	if err != nil {
		quoteUnavailable(c)
		return
	}
	if !resp.Success {
		quoteError(c, resp.Code)
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
		quoteUnavailable(c)
		return
	}
	if !resp.Success {
		quoteError(c, resp.Code)
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
	})
	if err != nil {
		quoteUnavailable(c)
		return
	}
	if !resp.Success {
		quoteError(c, resp.Code)
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
		quoteUnavailable(c)
		return
	}
	if !resp.Success {
		quoteError(c, resp.Code)
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
	})
	if err != nil {
		quoteUnavailable(c)
		return
	}
	if !resp.Success {
		quoteError(c, resp.Code)
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
		quoteUnavailable(c)
		return
	}
	if !resp.Success {
		quoteError(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
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
