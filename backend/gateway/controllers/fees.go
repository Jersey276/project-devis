package controllers

import (
	"log"
	"net/http"
	"os"

	quote "gateway/quote"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Fee error codes are shared with the quote service (same gRPC service).
var feeErrors = &serviceErrors{
	codes: map[int32]codeMapping{
		QuoteCodeNotFound:      {http.StatusNotFound, "Frais introuvable."},
		QuoteCodeAlreadyExists: {http.StatusConflict, "Ce frais existe déjà."},
		QuoteCodeInvalidInput:  {http.StatusBadRequest, "Données invalides."},
		QuoteCodeInternalError: {http.StatusInternalServerError, "Une erreur interne est survenue."},
	},
	unavailableMessage: "Service frais indisponible.",
}

// FeesRoutes wires the /fees API group against the quote gRPC service, which
// also hosts the premium fee catalog.
func FeesRoutes(r *gin.RouterGroup) {
	address := os.Getenv("QUOTE_SERVICE_ADDRESS")
	if address == "" {
		address = "localhost:50053"
	}
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to quote gRPC server: %v", err)
	}
	client := quote.NewQuoteServiceClient(conn)

	r.GET("", func(c *gin.Context) { ListFees(c, client) })
	r.POST("", func(c *gin.Context) { CreateFee(c, client) })
	r.GET("/:id", func(c *gin.Context) { GetFee(c, client) })
	r.PUT("/:id", func(c *gin.Context) { UpdateFee(c, client) })
	r.DELETE("/:id", func(c *gin.Context) { ArchiveFee(c, client) })
}

type feeInput struct {
	Category  string `json:"category" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Unit      string `json:"unit"`
	UnitPrice int64  `json:"unit_price"`
	TaxID     int32  `json:"tax_id"`
}

func marshalFee(f *quote.Fee) gin.H {
	if f == nil {
		return nil
	}
	return gin.H{
		"fee_id":     f.FeeId,
		"category":   f.Category,
		"name":       f.Name,
		"unit":       f.Unit,
		"unit_price": f.UnitPrice,
		"tax_id":     nullableInt(f.TaxId),
		"archived":   f.Archived,
	}
}

func ListFees(c *gin.Context, client quote.QuoteServiceClient) {
	resp, err := client.ListFees(c.Request.Context(), &quote.ListFeesRequest{
		UserId:          userIDFromCtx(c),
		IncludeArchived: c.Query("archived") == "true",
	})
	if err != nil {
		feeErrors.unavailable(c)
		return
	}
	if !resp.Success {
		feeErrors.reply(c, resp.Code)
		return
	}
	out := make([]gin.H, 0, len(resp.Fees))
	for _, f := range resp.Fees {
		out = append(out, marshalFee(f))
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "fees": out})
}

func CreateFee(c *gin.Context, client quote.QuoteServiceClient) {
	var input feeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.CreateFee(c.Request.Context(), &quote.CreateFeeRequest{
		UserId:    userIDFromCtx(c),
		Category:  input.Category,
		Name:      input.Name,
		Unit:      input.Unit,
		UnitPrice: input.UnitPrice,
		TaxId:     input.TaxID,
	})
	if err != nil {
		feeErrors.unavailable(c)
		return
	}
	if !resp.Success {
		if len(resp.ValidationErrors) > 0 {
			feeErrors.replyWithValidation(c, resp.Code, quoteValidationErrors(resp.ValidationErrors))
		} else {
			feeErrors.reply(c, resp.Code)
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "fee_id": resp.FeeId})
}

func GetFee(c *gin.Context, client quote.QuoteServiceClient) {
	resp, err := client.GetFee(c.Request.Context(), &quote.GetFeeRequest{
		FeeId:  c.Param("id"),
		UserId: userIDFromCtx(c),
	})
	if err != nil {
		feeErrors.unavailable(c)
		return
	}
	if !resp.Success {
		feeErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "fee": marshalFee(resp.Fee)})
}

func UpdateFee(c *gin.Context, client quote.QuoteServiceClient) {
	var input feeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.UpdateFee(c.Request.Context(), &quote.UpdateFeeRequest{
		FeeId:     c.Param("id"),
		UserId:    userIDFromCtx(c),
		Category:  input.Category,
		Name:      input.Name,
		Unit:      input.Unit,
		UnitPrice: input.UnitPrice,
		TaxId:     input.TaxID,
	})
	if err != nil {
		feeErrors.unavailable(c)
		return
	}
	if !resp.Success {
		if len(resp.ValidationErrors) > 0 {
			feeErrors.replyWithValidation(c, resp.Code, quoteValidationErrors(resp.ValidationErrors))
		} else {
			feeErrors.reply(c, resp.Code)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func ArchiveFee(c *gin.Context, client quote.QuoteServiceClient) {
	resp, err := client.ArchiveFee(c.Request.Context(), &quote.ArchiveFeeRequest{
		FeeId:  c.Param("id"),
		UserId: userIDFromCtx(c),
	})
	if err != nil {
		feeErrors.unavailable(c)
		return
	}
	if !resp.Success {
		feeErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
