package creditnote

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"

	"project-devis-export/internal/format"
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
	OssApplied       bool
	VatBreakdown     []vatView
	TotalHT          string
	TotalVAT         string
	TotalTTC         string
}

func Render(ctx context.Context, gt pdfConverter, cn *invoicepb.CreditNoteDetails) ([]byte, error) {
	html, err := renderHTML(cn)
	if err != nil {
		return nil, err
	}
	return gt.Convert(ctx, html)
}

// RenderPDFA3 renders the same credit note HTML as a PDF/A-3 (the base for a
// Factur-X document).
func RenderPDFA3(ctx context.Context, gt pdfConverter, cn *invoicepb.CreditNoteDetails) ([]byte, error) {
	html, err := renderHTML(cn)
	if err != nil {
		return nil, err
	}
	return gt.ConvertPDFA3(ctx, html)
}

func renderHTML(cn *invoicepb.CreditNoteDetails) ([]byte, error) {
	vm := buildViewModel(cn)
	var html bytes.Buffer
	if err := creditNoteTpl.Execute(&html, vm); err != nil {
		return nil, fmt.Errorf("render credit note template: %w", err)
	}
	return html.Bytes(), nil
}

func buildViewModel(cn *invoicepb.CreditNoteDetails) viewModel {
	// Amounts are stored positive; a credit note shows them as negatives.
	lines := make([]lineView, 0, len(cn.GetLines()))
	for _, l := range cn.GetLines() {
		lines = append(lines, lineView{
			Name:      l.GetName(),
			Quantity:  l.GetQuantity(),
			Unit:      l.GetUnit(),
			UnitPrice: format.Cents(l.GetUnitPriceCents()),
			TaxRate:   format.Rate(l.GetTaxRate()),
			Total:     format.Cents(-l.GetLineHtCents()),
		})
	}

	vat := make([]vatView, 0, len(cn.GetVatBreakdown()))
	for _, v := range cn.GetVatBreakdown() {
		vat = append(vat, vatView{
			Rate: format.Rate(v.GetTaxRate()),
			Base: format.Cents(-v.GetBaseHtCents()),
			VAT:  format.Cents(-v.GetVatCents()),
		})
	}

	return viewModel{
		CreditNoteNumber: cn.GetCreditNoteNumber(),
		InvoiceNumber:    cn.GetInvoiceNumber(),
		IssuedAt:         format.Date(cn.GetIssuedAt()),
		Reason:           cn.GetReason(),
		Issuer:           buildIssuer(cn.GetIssuer()),
		Recipient:        buildRecipient(cn.GetClient()),
		Lines:            lines,
		VatExempt:        cn.GetVatExempt(),
		OssApplied:       cn.GetOssApplied(),
		VatBreakdown:     vat,
		TotalHT:          format.Cents(-cn.GetTotalHtCents()),
		TotalVAT:         format.Cents(-cn.GetTotalVatCents()),
		TotalTTC:         format.Cents(-cn.GetTotalTtcCents()),
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
