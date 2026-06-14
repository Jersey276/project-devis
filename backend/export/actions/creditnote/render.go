package creditnote

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

var creditNoteTpl = template.Must(template.New("credit_note.html").Parse(string(templates.CreditNoteHTML)))

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
	CreditNoteNumber string
	InvoiceNumber    string
	IssuedAt         string
	Reason           string
	Issuer           partyView
	Recipient        partyView
	Lines            []lineView
	VatExempt        bool
	VatBreakdown     []vatView
	TotalHT          string
	TotalVAT         string
	TotalTTC         string
}

func Render(ctx context.Context, gt pdfConverter, cn *invoicepb.CreditNoteDetails) ([]byte, error) {
	vm := buildViewModel(cn)
	var html bytes.Buffer
	if err := creditNoteTpl.Execute(&html, vm); err != nil {
		return nil, fmt.Errorf("render credit note template: %w", err)
	}
	return gt.Convert(ctx, html.Bytes())
}

func buildViewModel(cn *invoicepb.CreditNoteDetails) viewModel {
	// Amounts are stored positive; a credit note shows them as negatives.
	lines := make([]lineView, 0, len(cn.GetLines()))
	for _, l := range cn.GetLines() {
		lines = append(lines, lineView{
			Name:      l.GetName(),
			Quantity:  l.GetQuantity(),
			Unit:      l.GetUnit(),
			UnitPrice: formatCents(l.GetUnitPriceCents()),
			TaxRate:   formatRate(l.GetTaxRate()),
			Total:     formatCents(-l.GetLineHtCents()),
		})
	}

	vat := make([]vatView, 0, len(cn.GetVatBreakdown()))
	for _, v := range cn.GetVatBreakdown() {
		vat = append(vat, vatView{
			Rate: formatRate(v.GetTaxRate()),
			Base: formatCents(-v.GetBaseHtCents()),
			VAT:  formatCents(-v.GetVatCents()),
		})
	}

	return viewModel{
		CreditNoteNumber: cn.GetCreditNoteNumber(),
		InvoiceNumber:    cn.GetInvoiceNumber(),
		IssuedAt:         formatDate(cn.GetIssuedAt()),
		Reason:           cn.GetReason(),
		Issuer:           buildIssuer(cn.GetIssuer()),
		Recipient:        buildRecipient(cn.GetClient()),
		Lines:            lines,
		VatExempt:        cn.GetVatExempt(),
		VatBreakdown:     vat,
		TotalHT:          formatCents(-cn.GetTotalHtCents()),
		TotalVAT:         formatCents(-cn.GetTotalVatCents()),
		TotalTTC:         formatCents(-cn.GetTotalTtcCents()),
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
	if i := strings.IndexByte(rfc3339, 'T'); i > 0 {
		return rfc3339[:i]
	}
	return rfc3339
}
