package controllers

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	invoice "gateway/invoice"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	InvoiceCodeNotFound            int32 = 1001
	InvoiceCodeAlreadyExists       int32 = 1002
	InvoiceCodeInvalidInput        int32 = 1003
	InvoiceCodeSourceNotEligible   int32 = 4001
	InvoiceCodeQuoteHasSchedule    int32 = 4002
	InvoiceCodeInvoiceFinalized    int32 = 4003
	InvoiceCodeMonthsAlreadyBilled int32 = 4004
	InvoiceCodeDependencyMissing   int32 = 4005

	InvoiceCodeCreditNoteLineAlreadyCredited int32 = 4006
	InvoiceCodeInvoiceNotIssued              int32 = 4007
	InvoiceCodeCreditNoteNoLinesLeft         int32 = 4008
	InvoiceCodeSealError                     int32 = 4009
	InvoiceCodeOSSDestinationTaxMissing      int32 = 4010
	InvoiceCodeLifecycleTransitionInvalid    int32 = 4011
	InvoiceCodeLifecycleRequiresIssued       int32 = 4012
	InvoiceCodePDPSubmissionFailed           int32 = 4013
	InvoiceCodeRecipientNotInDirectory       int32 = 4014
	InvoiceCodeReportSubmissionFailed        int32 = 4015

	InvoiceCodeInternalError int32 = 2001
)

func invoiceValidationErrors(errs []*invoice.ValidationError) []FieldError {
	out := make([]FieldError, len(errs))
	for i, e := range errs {
		out[i] = FieldError{Field: e.Field, Message: e.Message}
	}
	return out
}

var invoiceErrors = &serviceErrors{
	codes: map[int32]codeMapping{
		InvoiceCodeNotFound:                      {http.StatusNotFound, "Facture introuvable."},
		InvoiceCodeAlreadyExists:                 {http.StatusConflict, "Cette facture existe déjà."},
		InvoiceCodeInvalidInput:                  {http.StatusBadRequest, "Données invalides."},
		InvoiceCodeSourceNotEligible:             {http.StatusUnprocessableEntity, "La source n'est pas éligible à la facturation."},
		InvoiceCodeQuoteHasSchedule:              {http.StatusConflict, "Un échéancier existe pour ce devis : facturez depuis l'échéancier."},
		InvoiceCodeInvoiceFinalized:              {http.StatusConflict, "Cette facture est émise et ne peut plus être modifiée."},
		InvoiceCodeMonthsAlreadyBilled:           {http.StatusConflict, "Certains mois sélectionnés sont déjà facturés."},
		InvoiceCodeDependencyMissing:             {http.StatusUnprocessableEntity, "La facture fait référence à un client ou une adresse introuvable."},
		InvoiceCodeCreditNoteLineAlreadyCredited: {http.StatusConflict, "Une ou plusieurs lignes sélectionnées ont déjà fait l'objet d'un avoir."},
		InvoiceCodeInvoiceNotIssued:              {http.StatusConflict, "Seules les factures émises peuvent faire l'objet d'un avoir."},
		InvoiceCodeCreditNoteNoLinesLeft:         {http.StatusConflict, "Toutes les lignes de cette facture ont déjà été avoirées."},
		InvoiceCodeSealError:                     {http.StatusInternalServerError, "Une erreur interne est survenue."},
		InvoiceCodeOSSDestinationTaxMissing:      {http.StatusUnprocessableEntity, "Aucune TVA n'est configurée pour le pays du client (régime OSS). Configurez le taux du pays de destination avant d'émettre."},
		InvoiceCodeLifecycleTransitionInvalid:    {http.StatusConflict, "Transition de statut e-invoicing non autorisée."},
		InvoiceCodeLifecycleRequiresIssued:       {http.StatusConflict, "Le statut e-invoicing ne s'applique qu'aux factures émises."},
		InvoiceCodePDPSubmissionFailed:           {http.StatusBadGateway, "Le dépôt sur la plateforme a échoué. Réessayez plus tard."},
		InvoiceCodeRecipientNotInDirectory:       {http.StatusUnprocessableEntity, "Le destinataire est introuvable dans l'annuaire e-invoicing : vérifiez son SIRET."},
		InvoiceCodeReportSubmissionFailed:        {http.StatusBadGateway, "La transmission de l'e-reporting à la plateforme a échoué. Réessayez plus tard."},
		InvoiceCodeInternalError:                 {http.StatusInternalServerError, "Une erreur interne est survenue."},
	},
	unavailableMessage: "Service de facturation indisponible.",
}

// InvoicesRoutes wires the /invoices API group against the invoice gRPC service.
func InvoicesRoutes(r *gin.RouterGroup) {
	address := os.Getenv("INVOICE_SERVICE_ADDRESS")
	if address == "" {
		address = "localhost:50059"
	}
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect to invoice gRPC server: " + err.Error())
	}
	client := invoice.NewInvoiceServiceClient(conn)

	r.GET("", func(c *gin.Context) { ListInvoices(c, client) })
	r.GET("/verify-chain", func(c *gin.Context) { VerifyChain(c, client) })
	r.GET("/oss-status", func(c *gin.Context) { GetOSSThresholdStatus(c, client) })
	// E-reporting (B5/C5): period aggregates, not tied to one invoice.
	r.GET("/reports", func(c *gin.Context) { ListInvoiceReports(c, client) })
	r.POST("/reports", func(c *gin.Context) { SubmitInvoiceReport(c, client) })
	r.POST("/from-schedule", func(c *gin.Context) { CreateInvoiceFromSchedule(c, client) })
	r.POST("/from-quote", func(c *gin.Context) { CreateInvoiceFromQuote(c, client) })

	one := r.Group("/:id")
	one.GET("", func(c *gin.Context) { GetInvoice(c, client) })
	one.DELETE("", func(c *gin.Context) { DeleteDraftInvoice(c, client) })
	one.POST("/issue", func(c *gin.Context) { IssueInvoice(c, client) })
	one.POST("/paid", func(c *gin.Context) { MarkInvoicePaid(c, client) })
	one.POST("/credit-notes", func(c *gin.Context) { CreateCreditNote(c, client) })
	one.POST("/lifecycle", func(c *gin.Context) { SetInvoiceLifecycleStatus(c, client) })
	one.GET("/lifecycle-events", func(c *gin.Context) { ListInvoiceLifecycleEvents(c, client) })
	one.POST("/deposit", func(c *gin.Context) { DepositInvoice(c, client) })
}

func ListInvoices(c *gin.Context, client invoice.InvoiceServiceClient) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "20"), 10, 32)

	var statuses, lifecycleStatuses []string
	if raw := c.Query("statuses"); raw != "" {
		statuses = strings.Split(raw, ",")
	}
	if raw := c.Query("lifecycle_statuses"); raw != "" {
		lifecycleStatuses = strings.Split(raw, ",")
	}

	resp, err := client.ListInvoices(c.Request.Context(), &invoice.ListInvoicesRequest{
		UserId:   userIDFromCtx(c),
		QuoteId:  c.Query("quote_id"),
		Page:     int32(page),
		PageSize: int32(pageSize),
		Filters: &invoice.InvoiceFilters{
			Statuses:          statuses,
			LifecycleStatuses: lifecycleStatuses,
			IssuedFrom:        c.Query("issued_from"),
			IssuedTo:          c.Query("issued_to"),
			DueFrom:           c.Query("due_from"),
			DueTo:             c.Query("due_to"),
			ClientId:          c.Query("client_id"),
			QuoteIdFilter:     c.Query("quote_id_filter"),
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
	out := make([]gin.H, 0, len(resp.Invoices))
	for _, in := range resp.Invoices {
		out = append(out, gin.H{
			"invoice_id":       in.InvoiceId,
			"invoice_number":   in.InvoiceNumber,
			"status":           in.Status,
			"quote_id":         in.QuoteId,
			"schedule_id":      in.ScheduleId,
			"issued_at":        in.IssuedAt,
			"due_date":         in.DueDate,
			"total_ttc_cents":  in.TotalTtcCents,
			"lifecycle_status": in.LifecycleStatus,
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "invoices": out, "total": resp.Total})
}

func CreateInvoiceFromSchedule(c *gin.Context, client invoice.InvoiceServiceClient) {
	var input struct {
		ScheduleID   string  `json:"schedule_id" binding:"required"`
		MonthIndexes []int32 `json:"month_indexes" binding:"required"`
		SaleDate     string  `json:"sale_date"`
		DueInDays    int32   `json:"due_in_days"`
		IssueNow     bool    `json:"issue_now"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.CreateInvoiceFromSchedule(c.Request.Context(), &invoice.CreateInvoiceFromScheduleRequest{
		UserId:       userIDFromCtx(c),
		ScheduleId:   input.ScheduleID,
		MonthIndexes: input.MonthIndexes,
		SaleDate:     input.SaleDate,
		DueInDays:    input.DueInDays,
		IssueNow:     input.IssueNow,
	})
	replyCreate(c, resp, err)
}

func CreateInvoiceFromQuote(c *gin.Context, client invoice.InvoiceServiceClient) {
	var input struct {
		QuoteID   string `json:"quote_id" binding:"required"`
		SaleDate  string `json:"sale_date"`
		DueInDays int32  `json:"due_in_days"`
		IssueNow  bool   `json:"issue_now"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.CreateInvoiceFromQuote(c.Request.Context(), &invoice.CreateInvoiceFromQuoteRequest{
		UserId:    userIDFromCtx(c),
		QuoteId:   input.QuoteID,
		SaleDate:  input.SaleDate,
		DueInDays: input.DueInDays,
		IssueNow:  input.IssueNow,
	})
	replyCreate(c, resp, err)
}

func IssueInvoice(c *gin.Context, client invoice.InvoiceServiceClient) {
	resp, err := client.IssueInvoice(c.Request.Context(), &invoice.IssueInvoiceRequest{
		InvoiceId: c.Param("id"),
		UserId:    userIDFromCtx(c),
	})
	replyCreate(c, resp, err)
}

func GetInvoice(c *gin.Context, client invoice.InvoiceServiceClient) {
	resp, err := client.GetInvoice(c.Request.Context(), &invoice.GetInvoiceRequest{
		InvoiceId: c.Param("id"),
		UserId:    userIDFromCtx(c),
	})
	if err != nil {
		invoiceErrors.unavailable(c)
		return
	}
	if !resp.Success {
		invoiceErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "invoice": invoiceDetailsToJSON(resp.Invoice)})
}

func CreateCreditNote(c *gin.Context, client invoice.InvoiceServiceClient) {
	var input struct {
		Positions []int32 `json:"positions"`
		Reason    string  `json:"reason"`
	}
	// positions optional (empty = total credit); bind is tolerant of empty body.
	_ = c.ShouldBindJSON(&input)

	resp, err := client.CreateCreditNote(c.Request.Context(), &invoice.CreateCreditNoteRequest{
		UserId:    userIDFromCtx(c),
		InvoiceId: c.Param("id"),
		Positions: input.Positions,
		Reason:    input.Reason,
	})
	if err != nil {
		invoiceErrors.unavailable(c)
		return
	}
	if !resp.Success {
		if len(resp.ValidationErrors) > 0 {
			invoiceErrors.replyWithValidation(c, resp.Code, invoiceValidationErrors(resp.ValidationErrors))
		} else {
			invoiceErrors.reply(c, resp.Code)
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"success":            true,
		"credit_note_id":     resp.CreditNoteId,
		"credit_note_number": resp.CreditNoteNumber,
	})
}

func VerifyChain(c *gin.Context, client invoice.InvoiceServiceClient) {
	resp, err := client.VerifyChain(c.Request.Context(), &invoice.VerifyChainRequest{
		UserId: userIDFromCtx(c),
	})
	if err != nil {
		invoiceErrors.unavailable(c)
		return
	}
	if !resp.Success {
		invoiceErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"ok":              resp.Ok,
		"checked":         resp.Checked,
		"broken_doc_id":   resp.BrokenDocId,
		"broken_doc_type": resp.BrokenDocType,
		"broken_index":    resp.BrokenIndex,
		"reason":          resp.Reason,
	})
}

func GetOSSThresholdStatus(c *gin.Context, client invoice.InvoiceServiceClient) {
	resp, err := client.GetOSSThresholdStatus(c.Request.Context(), &invoice.GetOSSThresholdStatusRequest{
		UserId: userIDFromCtx(c),
	})
	if err != nil {
		invoiceErrors.unavailable(c)
		return
	}
	if !resp.Success {
		invoiceErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":             true,
		"year":                resp.Year,
		"cumulative_ht_cents": resp.CumulativeHtCents,
		"threshold_cents":     resp.ThresholdCents,
		"oss_enabled":         resp.OssEnabled,
		"oss_active":          resp.OssActive,
		"prior_year_over_threshold":      resp.PriorYearOverThreshold,
		"prior_year_cumulative_ht_cents": resp.PriorYearCumulativeHtCents,
	})
}

func MarkInvoicePaid(c *gin.Context, client invoice.InvoiceServiceClient) {
	resp, err := client.MarkInvoicePaid(c.Request.Context(), &invoice.MarkInvoicePaidRequest{
		InvoiceId: c.Param("id"),
		UserId:    userIDFromCtx(c),
	})
	replyGeneric(c, resp, err)
}

func DeleteDraftInvoice(c *gin.Context, client invoice.InvoiceServiceClient) {
	resp, err := client.DeleteDraftInvoice(c.Request.Context(), &invoice.DeleteDraftInvoiceRequest{
		InvoiceId: c.Param("id"),
		UserId:    userIDFromCtx(c),
	})
	replyGeneric(c, resp, err)
}

func SetInvoiceLifecycleStatus(c *gin.Context, client invoice.InvoiceServiceClient) {
	var input struct {
		Status string `json:"status" binding:"required"`
		Note   string `json:"note"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.SetInvoiceLifecycleStatus(c.Request.Context(), &invoice.SetInvoiceLifecycleStatusRequest{
		InvoiceId: c.Param("id"),
		UserId:    userIDFromCtx(c),
		Status:    input.Status,
		Note:      input.Note,
	})
	replyGeneric(c, resp, err)
}

func DepositInvoice(c *gin.Context, client invoice.InvoiceServiceClient) {
	var input struct {
		Note string `json:"note"`
	}
	_ = c.ShouldBindJSON(&input) // body optional (note only)
	resp, err := client.DepositInvoice(c.Request.Context(), &invoice.DepositInvoiceRequest{
		InvoiceId: c.Param("id"),
		UserId:    userIDFromCtx(c),
		Note:      input.Note,
	})
	replyGeneric(c, resp, err)
}

func ListInvoiceLifecycleEvents(c *gin.Context, client invoice.InvoiceServiceClient) {
	resp, err := client.ListInvoiceLifecycleEvents(c.Request.Context(), &invoice.ListInvoiceLifecycleEventsRequest{
		InvoiceId: c.Param("id"),
		UserId:    userIDFromCtx(c),
	})
	if err != nil {
		invoiceErrors.unavailable(c)
		return
	}
	if !resp.Success {
		invoiceErrors.reply(c, resp.Code)
		return
	}
	out := make([]gin.H, 0, len(resp.Events))
	for _, e := range resp.Events {
		out = append(out, gin.H{
			"status":     e.Status,
			"note":       e.Note,
			"created_at": e.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "events": out})
}

func SubmitInvoiceReport(c *gin.Context, client invoice.InvoiceServiceClient) {
	var input struct {
		Kind  string `json:"kind" binding:"required"`
		Year  int32  `json:"year" binding:"required"`
		Month int32  `json:"month" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.SubmitInvoiceReport(c.Request.Context(), &invoice.SubmitInvoiceReportRequest{
		UserId: userIDFromCtx(c),
		Kind:   input.Kind,
		Year:   input.Year,
		Month:  input.Month,
	})
	replyGeneric(c, resp, err)
}

func ListInvoiceReports(c *gin.Context, client invoice.InvoiceServiceClient) {
	resp, err := client.ListInvoiceReports(c.Request.Context(), &invoice.ListInvoiceReportsRequest{
		UserId: userIDFromCtx(c),
		Kind:   c.Query("kind"),
	})
	if err != nil {
		invoiceErrors.unavailable(c)
		return
	}
	if !resp.Success {
		invoiceErrors.reply(c, resp.Code)
		return
	}
	out := make([]gin.H, 0, len(resp.Reports))
	for _, r := range resp.Reports {
		out = append(out, gin.H{
			"kind":            r.Kind,
			"year":            r.Year,
			"month":           r.Month,
			"status":          r.Status,
			"total_ht_cents":  r.TotalHtCents,
			"total_vat_cents": r.TotalVatCents,
			"submitted_at":    r.SubmittedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "reports": out})
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func replyCreate(c *gin.Context, resp *invoice.CreateInvoiceResponse, err error) {
	if err != nil {
		invoiceErrors.unavailable(c)
		return
	}
	if !resp.Success {
		if len(resp.ValidationErrors) > 0 {
			invoiceErrors.replyWithValidation(c, resp.Code, invoiceValidationErrors(resp.ValidationErrors))
		} else {
			invoiceErrors.reply(c, resp.Code)
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"success":        true,
		"invoice_id":     resp.InvoiceId,
		"invoice_number": resp.InvoiceNumber,
	})
}

func replyGeneric(c *gin.Context, resp *invoice.GenericResponse, err error) {
	if err != nil {
		invoiceErrors.unavailable(c)
		return
	}
	if !resp.Success {
		invoiceErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func invoicePartyToJSON(p *invoice.InvoiceParty) gin.H {
	if p == nil {
		return gin.H{}
	}
	return gin.H{
		"company":           p.Company,
		"first_name":        p.FirstName,
		"last_name":         p.LastName,
		"siren":             p.Siren,
		"siret":             p.Siret,
		"vat":               p.Vat,
		"email":             p.Email,
		"phone":             p.Phone,
		"logo_url":          p.LogoUrl,
		"street":            p.Street,
		"additional_street": p.AdditionalStreet,
		"zip_code":          p.ZipCode,
		"city":              p.City,
		"country_code":      p.CountryCode,
		"iban":              p.Iban,
		"bic":               p.Bic,
	}
}

func invoiceDetailsToJSON(d *invoice.InvoiceDetails) gin.H {
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
		"invoice_id":           d.InvoiceId,
		"quote_id":             d.QuoteId,
		"schedule_id":          d.ScheduleId,
		"billed_month_indexes": d.BilledMonthIndexes,
		"status":               d.Status,
		"invoice_number":       d.InvoiceNumber,
		"issued_at":            d.IssuedAt,
		"sale_date":            d.SaleDate,
		"due_date":             d.DueDate,
		"issuer":               invoicePartyToJSON(d.Issuer),
		"client":               invoicePartyToJSON(d.Client),
		"lines":                lines,
		"vat_breakdown":        vat,
		"total_ht_cents":       d.TotalHtCents,
		"total_vat_cents":      d.TotalVatCents,
		"total_ttc_cents":      d.TotalTtcCents,
		"vat_exempt":           d.VatExempt,
		"oss_applied":          d.OssApplied,
		"credited_positions":   d.CreditedPositions,
		"lifecycle_status":     d.LifecycleStatus,
	}
}
