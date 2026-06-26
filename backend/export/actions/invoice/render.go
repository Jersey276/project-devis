package invoice

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
	OssApplied    bool
	VatBreakdown  []vatView
	TotalHT       string
	TotalVAT      string
	TotalTTC      string
}

func Render(ctx context.Context, gt pdfConverter, in *invoicepb.InvoiceDetails) ([]byte, error) {
	html, err := renderHTML(in)
	if err != nil {
		return nil, err
	}
	return gt.Convert(ctx, html)
}

// RenderPDFA3 renders the same invoice HTML but as a PDF/A-3 (the base for a
// Factur-X document).
func RenderPDFA3(ctx context.Context, gt pdfConverter, in *invoicepb.InvoiceDetails) ([]byte, error) {
	html, err := renderHTML(in)
	if err != nil {
		return nil, err
	}
	return gt.ConvertPDFA3(ctx, html)
}

func renderHTML(in *invoicepb.InvoiceDetails) ([]byte, error) {
	vm := buildViewModel(in)
	var html bytes.Buffer
	if err := invoiceTpl.Execute(&html, vm); err != nil {
		return nil, fmt.Errorf("render invoice template: %w", err)
	}
	return html.Bytes(), nil
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
			UnitPrice: format.Cents(l.GetUnitPriceCents()),
			TaxRate:   format.Rate(l.GetTaxRate()),
			Total:     format.Cents(l.GetLineHtCents()),
		})
	}

	vat := make([]vatView, 0, len(in.GetVatBreakdown()))
	for _, v := range in.GetVatBreakdown() {
		vat = append(vat, vatView{
			Rate: format.Rate(v.GetTaxRate()),
			Base: format.Cents(v.GetBaseHtCents()),
			VAT:  format.Cents(v.GetVatCents()),
		})
	}

	return viewModel{
		Title:         title,
		InvoiceNumber: in.GetInvoiceNumber(),
		IsDraft:       draft,
		Status:        in.GetStatus(),
		IssuedAt:      format.Date(in.GetIssuedAt()),
		SaleDate:      in.GetSaleDate(),
		DueDate:       in.GetDueDate(),
		Issuer:        buildIssuer(in.GetIssuer()),
		Recipient:     buildRecipient(in.GetClient()),
		Lines:         lines,
		VatExempt:     in.GetVatExempt(),
		OssApplied:    in.GetOssApplied(),
		VatBreakdown:  vat,
		TotalHT:       format.Cents(in.GetTotalHtCents()),
		TotalVAT:      format.Cents(in.GetTotalVatCents()),
		TotalTTC:      format.Cents(in.GetTotalTtcCents()),
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

