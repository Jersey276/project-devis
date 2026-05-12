package controllers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	export "gateway/export"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	ExportCodeNotFound      int32 = 3001
	ExportCodeForbidden     int32 = 3002
	ExportCodeInternalError int32 = 3003
	ExportCodeInvalidInput  int32 = 3004
)

var exportErrorMap = map[int32]struct {
	Status  int
	Message string
}{
	ExportCodeNotFound:      {http.StatusNotFound, "Devis introuvable."},
	ExportCodeForbidden:     {http.StatusForbidden, "Accès refusé."},
	ExportCodeInvalidInput:  {http.StatusBadRequest, "Requête invalide."},
	ExportCodeInternalError: {http.StatusInternalServerError, "Une erreur interne est survenue."},
}

func exportError(c *gin.Context, code int32) {
	if mapped, ok := exportErrorMap[code]; ok {
		c.JSON(mapped.Status, gin.H{"success": false, "message": mapped.Message, "code": code})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Une erreur inconnue est survenue.", "code": code})
	}
}

func exportUnavailable(c *gin.Context) {
	c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": "Service export indisponible."})
}

// ExportRoutes wires the /export API group against the export gRPC service.
func ExportRoutes(r *gin.RouterGroup) {
	address := os.Getenv("EXPORT_SERVICE_ADDRESS")
	if address == "" {
		address = "localhost:50054"
	}
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to export gRPC server: %v", err)
	}
	client := export.NewExportServiceClient(conn)

	r.GET("/quotes/:id", func(c *gin.Context) { ExportQuote(c, client) })
}

func ExportQuote(c *gin.Context, client export.ExportServiceClient) {
	resp, err := client.ExportQuote(c.Request.Context(), &export.ExportQuoteRequest{
		QuoteId: c.Param("id"),
		UserId:  userIDFromCtx(c),
	})
	if err != nil {
		exportUnavailable(c)
		return
	}
	if !resp.Success {
		exportError(c, resp.Code)
		return
	}
	c.Header("Content-Disposition", contentDispositionAttachment(resp.Filename))
	c.Data(http.StatusOK, "application/pdf", resp.Pdf)
}

// contentDispositionAttachment builds a header value compatible with both
// legacy clients (filename="…" with non-ASCII stripped) and modern browsers
// (filename*=UTF-8'' with percent-encoded UTF-8) per RFC 5987 / 6266.
func contentDispositionAttachment(filename string) string {
	ascii := stripNonASCII(filename)
	if ascii == "" {
		ascii = "download.pdf"
	}
	return fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`,
		ascii, url.PathEscape(filename))
}

func stripNonASCII(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r < 0x80 && r != '"' && r != '\\' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
