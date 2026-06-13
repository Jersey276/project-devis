package invoice

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	invoicepb "project-devis-export/services/invoice"
	"project-devis-export/templates"
)

var invoiceTpl = template.Must(template.New("invoice.html").Parse(string(templates.InvoiceHTML)))

type partyView struct {
	LogoURL string
	Title   string
	Lines   []string
}

type lineView struct {
	Name      string
	Quantity  string
	Unit      string
	UnitPrice string
	TaxRate   string
	Total     string
}

type vatView struct {
	Rate string
	Base string
	VAT  string
}

type viewModel struct {
	Title         string
	InvoiceNumber string
	IsDraft       bool
	Status        string
	IssuedAt      string
	SaleDate      string
	DueDate       string
	Issuer        partyView
	Recipient     partyView
	Lines         []lineView
	VatExempt     bool
	VatBreakdown  []vatView
	TotalHT       string
	TotalVAT      string
	TotalTTC      string
}

func Render(ctx context.Context, gt pdfConverter, in *invoicepb.InvoiceDetails) ([]byte, error) {
	vm := buildViewModel(in)
	var html bytes.Buffer
	if err := invoiceTpl.Execute(&html, vm); err != nil {
		return nil, fmt.Errorf("render invoice template: %w", err)
	}
	return gt.Convert(ctx, html.Bytes())
}

func buildViewModel(in *invoicepb.InvoiceDetails) viewModel {
	draft := in.GetStatus() == "DRAFT"
	title := "FACTURE"
	if draft {
		title = "FACTURE (BROUILLON)"
	}

	lines := make([]lineView, 0, len(in.GetLines()))
	for _, l := range in.GetLines() {
		lines = append(lines, lineView{
			Name:      l.GetName(),
			Quantity:  l.GetQuantity(),
			Unit:      l.GetUnit(),
			UnitPrice: formatCents(l.GetUnitPriceCents()),
			TaxRate:   formatRate(l.GetTaxRate()),
			Total:     formatCents(l.GetLineHtCents()),
		})
	}

	vat := make([]vatView, 0, len(in.GetVatBreakdown()))
	for _, v := range in.GetVatBreakdown() {
		vat = append(vat, vatView{
			Rate: formatRate(v.GetTaxRate()),
			Base: formatCents(v.GetBaseHtCents()),
			VAT:  formatCents(v.GetVatCents()),
		})
	}

	return viewModel{
		Title:         title,
		InvoiceNumber: in.GetInvoiceNumber(),
		IsDraft:       draft,
		Status:        in.GetStatus(),
		IssuedAt:      formatDate(in.GetIssuedAt()),
		SaleDate:      in.GetSaleDate(),
		DueDate:       in.GetDueDate(),
		Issuer:        buildIssuer(in.GetIssuer()),
		Recipient:     buildRecipient(in.GetClient()),
		Lines:         lines,
		VatExempt:     in.GetVatExempt(),
		VatBreakdown:  vat,
		TotalHT:       formatCents(in.GetTotalHtCents()),
		TotalVAT:      formatCents(in.GetTotalVatCents()),
		TotalTTC:      formatCents(in.GetTotalTtcCents()),
	}
}

func buildIssuer(p *invoicepb.InvoiceParty) partyView {
	v := partyView{}
	if p == nil {
		return v
	}
	v.LogoURL = p.GetLogoUrl()
	v.Title = p.GetCompany()
	v.Lines = appendAddressLines(v.Lines, p)
	if p.GetEmail() != "" {
		v.Lines = append(v.Lines, p.GetEmail())
	}
	if p.GetPhone() != "" {
		v.Lines = append(v.Lines, p.GetPhone())
	}
	if p.GetSiren() != "" {
		v.Lines = append(v.Lines, "SIREN : "+p.GetSiren())
	}
	if p.GetVat() != "" {
		v.Lines = append(v.Lines, "TVA : "+p.GetVat())
	}
	return v
}

func buildRecipient(p *invoicepb.InvoiceParty) partyView {
	v := partyView{}
	if p == nil {
		return v
	}
	v.Title = strings.TrimSpace(p.GetFirstName() + " " + p.GetLastName())
	if p.GetCompany() != "" {
		v.Lines = append(v.Lines, p.GetCompany())
	}
	v.Lines = appendAddressLines(v.Lines, p)
	if p.GetEmail() != "" {
		v.Lines = append(v.Lines, p.GetEmail())
	}
	return v
}

func appendAddressLines(dst []string, p *invoicepb.InvoiceParty) []string {
	if p.GetStreet() != "" {
		dst = append(dst, p.GetStreet())
	}
	if p.GetAdditionalStreet() != "" {
		dst = append(dst, p.GetAdditionalStreet())
	}
	if cityLine := strings.TrimSpace(p.GetZipCode() + " " + p.GetCity()); cityLine != "" {
		dst = append(dst, cityLine)
	}
	return dst
}

func formatCents(cents int64) string {
	neg := cents < 0
	if neg {
		cents = -cents
	}
	euros := cents / 100
	rem := cents % 100
	euroStr := groupThousands(strconv.FormatInt(euros, 10))
	sign := ""
	if neg {
		sign = "-"
	}
	return fmt.Sprintf("%s%s,%02d €", sign, euroStr, rem)
}

func groupThousands(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}
	var b strings.Builder
	pre := n % 3
	if pre > 0 {
		b.WriteString(s[:pre])
		if n > pre {
			b.WriteByte(' ')
		}
	}
	for i := pre; i < n; i += 3 {
		b.WriteString(s[i : i+3])
		if i+3 < n {
			b.WriteByte(' ')
		}
	}
	return b.String()
}

// formatRate renders a tax-rate string as a percentage label, e.g. "20" → "20 %".
func formatRate(rate string) string {
	rate = strings.TrimSpace(rate)
	if rate == "" {
		return "0 %"
	}
	return rate + " %"
}

func formatDate(rfc3339 string) string {
	if rfc3339 == "" {
		return ""
	}
	// The value is RFC3339; keep just the date part for display.
	if i := strings.IndexByte(rfc3339, 'T'); i > 0 {
		return rfc3339[:i]
	}
	return rfc3339
}
