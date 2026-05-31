package controllers

import (
	"fmt"
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
	ExportCodeNotFound          int32 = 3001
	ExportCodeInternalError     int32 = 3003
	ExportCodeInvalidInput      int32 = 3004
	ExportCodeDependencyMissing int32 = 3005

	// 8 MiB — generous headroom for realistic quote PDFs (typical: 50 KiB–1 MiB).
	// If we start embedding heavy media we'll switch to server-streaming gRPC
	// rather than raise this further. Mirrored in backend/export/main.go.
	maxExportMessageBytes = 8 * 1024 * 1024
)

var exportErrors = &serviceErrors{
	codes: map[int32]codeMapping{
		ExportCodeNotFound:          {http.StatusNotFound, "Devis introuvable."},
		ExportCodeInvalidInput:      {http.StatusBadRequest, "Requête invalide."},
		ExportCodeInternalError:     {http.StatusInternalServerError, "Une erreur interne est survenue."},
		ExportCodeDependencyMissing: {http.StatusUnprocessableEntity, "Le devis fait référence à un client ou une adresse introuvable."},
	},
	unavailableMessage: "Service export indisponible.",
}

// ExportRoutes wires the /export API group against the export gRPC service.
func ExportRoutes(r *gin.RouterGroup) {
	exportAddress := os.Getenv("EXPORT_SERVICE_ADDRESS")
	if exportAddress == "" {
		exportAddress = "localhost:50054"
	}
	exportConn, err := grpc.NewClient(exportAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxExportMessageBytes),
			grpc.MaxCallSendMsgSize(maxExportMessageBytes),
		),
	)
	if err != nil {
		panic("failed to connect to export gRPC server: " + err.Error())
	}
	exportClient := export.NewExportServiceClient(exportConn)

	r.GET("/quotes/:id", func(c *gin.Context) { ExportQuote(c, exportClient) })
	r.GET("/schedules/:id", func(c *gin.Context) { ExportSchedule(c, exportClient) })
}

func ExportQuote(c *gin.Context, client export.ExportServiceClient) {
	resp, err := client.ExportQuote(c.Request.Context(), &export.ExportQuoteRequest{
		QuoteId: c.Param("id"),
		UserId:  userIDFromCtx(c),
	})
	if err != nil {
		exportErrors.unavailable(c)
		return
	}
	if !resp.Success {
		exportErrors.reply(c, resp.Code)
		return
	}
	c.Header("Content-Disposition", contentDispositionAttachment(resp.Filename))
	c.Data(http.StatusOK, "application/pdf", resp.Pdf)
}

func ExportSchedule(c *gin.Context, exportClient export.ExportServiceClient) {
	scheduleID := c.Param("id")
	exportResp, err := exportClient.ExportSchedule(c.Request.Context(), &export.ExportScheduleRequest{
		ScheduleId: scheduleID,
		UserId:     userIDFromCtx(c),
	})
	if err != nil {
		exportErrors.unavailable(c)
		return
	}
	if !exportResp.Success {
		exportErrors.reply(c, exportResp.Code)
		return
	}

	filename := exportResp.Filename
	if strings.TrimSpace(filename) == "" {
		filename = fmt.Sprintf("echeancier-%s.pdf", scheduleID)
	}
	c.Header("Content-Disposition", contentDispositionAttachment(filename))
	c.Data(http.StatusOK, "application/pdf", exportResp.Pdf)
}

// Emits both the legacy `filename="…"` (non-ASCII stripped) and the
// `filename*=UTF-8”` form (RFC 5987 / 6266) so accented filenames round-trip
// through older clients without being mangled.
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
