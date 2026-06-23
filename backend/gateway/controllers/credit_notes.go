package controllers

import (
	"net/http"
	"os"
	"strconv"

	invoice "gateway/invoice"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CreditNotesRoutes wires the /credit-notes API group against the invoice gRPC
// service (credit notes live in the same service as invoices). Creation is
// exposed under /invoices/:id/credit-notes (see InvoicesRoutes); this group
// serves reads.
func CreditNotesRoutes(r *gin.RouterGroup) {
	address := os.Getenv("INVOICE_SERVICE_ADDRESS")
	if address == "" {
		address = "localhost:50059"
	}
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect to invoice gRPC server: " + err.Error())
	}
	client := invoice.NewInvoiceServiceClient(conn)

	r.GET("", func(c *gin.Context) { ListCreditNotes(c, client) })
	r.GET("/:id", func(c *gin.Context) { GetCreditNote(c, client) })
}

func ListCreditNotes(c *gin.Context, client invoice.InvoiceServiceClient) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "20"), 10, 32)

	resp, err := client.ListCreditNotes(c.Request.Context(), &invoice.ListCreditNotesRequest{
		UserId:    userIDFromCtx(c),
		InvoiceId: c.Query("invoice_id"),
		Page:      int32(page),
		PageSize:  int32(pageSize),
		Filters: &invoice.CreditNoteFilters{
			IsTotal:    c.Query("is_total"),
			IssuedFrom: c.Query("issued_from"),
			IssuedTo:   c.Query("issued_to"),
		},
	})
	if err != nil {
		invoiceErrors.unavailable(c)
		return
	}
	if !resp.Success {
		invoiceErrors.reply(c, resp.Code)
		return
	}
	out := make([]gin.H, 0, len(resp.CreditNotes))
	for _, cn := range resp.CreditNotes {
		out = append(out, creditNoteSummaryToJSON(cn))
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "credit_notes": out, "total": resp.Total})
}

func GetCreditNote(c *gin.Context, client invoice.InvoiceServiceClient) {
	resp, err := client.GetCreditNote(c.Request.Context(), &invoice.GetCreditNoteRequest{
		CreditNoteId: c.Param("id"),
		UserId:       userIDFromCtx(c),
	})
	if err != nil {
		invoiceErrors.unavailable(c)
		return
	}
	if !resp.Success {
		invoiceErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "credit_note": creditNoteDetailsToJSON(resp.CreditNote)})
}

func creditNoteSummaryToJSON(cn *invoice.CreditNoteSummary) gin.H {
	return gin.H{
		"credit_note_id":     cn.CreditNoteId,
		"credit_note_number": cn.CreditNoteNumber,
		"invoice_id":         cn.InvoiceId,
		"invoice_number":     cn.InvoiceNumber,
		"issued_at":          cn.IssuedAt,
		"is_total":           cn.IsTotal,
		"total_ttc_cents":    cn.TotalTtcCents,
	}
}

func creditNoteDetailsToJSON(d *invoice.CreditNoteDetails) gin.H {
	if d == nil {
		return gin.H{}
	}
	lines := make([]gin.H, 0, len(d.Lines))
	for _, l := range d.Lines {
		lines = append(lines, gin.H{
			"quote_line_id":    l.QuoteLineId,
			"name":             l.Name,
			"unit":             l.Unit,
			"quantity":         l.Quantity,
			"unit_price_cents": l.UnitPriceCents,
			"line_ht_cents":    l.LineHtCents,
			"tax_id":           l.TaxId,
			"tax_rate":         l.TaxRate,
			"tax_label":        l.TaxLabel,
		})
	}
	vat := make([]gin.H, 0, len(d.VatBreakdown))
	for _, v := range d.VatBreakdown {
		vat = append(vat, gin.H{
			"tax_rate":      v.TaxRate,
			"base_ht_cents": v.BaseHtCents,
			"vat_cents":     v.VatCents,
		})
	}
	return gin.H{
		"credit_note_id":     d.CreditNoteId,
		"invoice_id":         d.InvoiceId,
		"invoice_number":     d.InvoiceNumber,
		"credit_note_number": d.CreditNoteNumber,
		"issued_at":          d.IssuedAt,
		"reason":             d.Reason,
		"is_total":           d.IsTotal,
		"issuer":             invoicePartyToJSON(d.Issuer),
		"client":             invoicePartyToJSON(d.Client),
		"lines":              lines,
		"vat_breakdown":      vat,
		"total_ht_cents":     d.TotalHtCents,
		"total_vat_cents":    d.TotalVatCents,
		"total_ttc_cents":    d.TotalTtcCents,
		"vat_exempt":         d.VatExempt,
	}
}
