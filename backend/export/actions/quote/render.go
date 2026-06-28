package quote

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"project-devis-export/internal/format"
	"project-devis-export/quote"
	"project-devis-export/templates"
	"project-devis-export/users"
)

type vatView struct {
	Rate string
	Base string
	VAT  string
}

// rowView is a polymorphic row driven by Kind:
//
//	"line"     — standard billable line
//	"subline"  — indented child of a group or detailed line
//	"group"    — section header (colspan, no numeric columns)
//	"text"     — free-text row (colspan, no numeric columns)
//	"detailed" — parent line with Sublines expanded below it
type rowView struct {
	Kind      string
	Name      string
	Quantity  string
	Unit      string
	UnitPrice string
	TaxRate   string
	Total     string
	Sublines  []rowView
}

var quoteTpl = template.Must(template.New("quote.html").Parse(string(templates.QuoteHTML)))

type renderInput struct {
	Quote         *quote.Quote
	Lines         []*quote.QuoteLine
	User          *users.User
	UserAddress   *users.Address
	Client        *users.Client
	ClientAddress *users.Address
	Taxes         map[int32]*users.Tax
}

type viewModel struct {
	ShortID                 string
	QuoteName               string
	IssuedAt                string
	ValidUntil              string
	PaymentTerms            string
	Sender                  partyView
	Recipient               partyView
	Rows                    []rowView
	OptionRows              []rowView
	OptionTotalHT           string
	VatBreakdown            []vatView
	TotalHT                 string
	TotalVAT                string
	TotalTTC                string
	SenderSignatureLabel    string
	RecipientSignatureLabel string
}

type partyView struct {
	Title string
	Lines []string
}

// lineData mirrors the QuoteLineData TypeScript type stored as JSON in QuoteLine.Data.
type lineData struct {
	Kind         string        `json:"kind"`
	Description  string        `json:"description"`
	Option       bool          `json:"option"`
	ParentLineID string        `json:"parent_line_id"`
	FeeID        string        `json:"fee_id"`
	Sublines     []sublineData `json:"sublines"`
}

type sublineData struct {
	Name      string `json:"name"`
	Quantity  string `json:"quantity"`
	Unit      string `json:"unit"`
	UnitPrice int64  `json:"unit_price"` // cents
	Option    bool   `json:"option"`
	FeeID     string `json:"fee_id"`
}

func parseLineData(raw string) lineData {
	if raw == "" || raw == "{}" {
		return lineData{}
	}
	var d lineData
	_ = json.Unmarshal([]byte(raw), &d)
	return d
}

func Render(ctx context.Context, gt pdfConverter, in renderInput) ([]byte, error) {
	vm := buildViewModel(in)

	var html bytes.Buffer
	if err := quoteTpl.Execute(&html, vm); err != nil {
		return nil, fmt.Errorf("render template: %w", err)
	}
	return gt.Convert(ctx, html.Bytes())
}

func buildViewModel(in renderInput) viewModel {
	type vatAccum struct {
		baseHT  int64
		vatCent int64
	}
	vatByRate := map[string]*vatAccum{}
	var rateOrder []string

	totalHT := int64(0)
	totalVAT := int64(0)
	optionTotalHT := int64(0)

	accumulateVAT := func(lineHT int64, taxRateStr string) {
		if taxRateStr == "" {
			return
		}
		rateF, err := strconv.ParseFloat(strings.ReplaceAll(taxRateStr, ",", "."), 64)
		if err != nil || rateF <= 0 {
			return
		}
		lineVAT := int64(float64(lineHT) * rateF / 100)
		totalVAT += lineVAT
		if _, exists := vatByRate[taxRateStr]; !exists {
			vatByRate[taxRateStr] = &vatAccum{}
			rateOrder = append(rateOrder, taxRateStr)
		}
		vatByRate[taxRateStr].baseHT += lineHT
		vatByRate[taxRateStr].vatCent += lineVAT
	}

	taxRateFor := func(l *quote.QuoteLine) string {
		if l.TaxId == 0 {
			return ""
		}
		if tax, ok := in.Taxes[l.TaxId]; ok {
			return tax.Rate
		}
		return ""
	}

	rows := make([]rowView, 0, len(in.Lines))
	var optionRows []rowView

	// appendLeaf handles both "line" and "subline" kinds — same logic, different Kind value.
	appendLeaf := func(kind string, l *quote.QuoteLine, isOption bool) {
		qty := parseQuantity(l.Quantity)
		lineHT := int64(qty * float64(l.UnitPrice))
		taxRateStr := taxRateFor(l)
		taxRateDisplay := "--"
		if taxRateStr != "" {
			taxRateDisplay = format.Rate(taxRateStr)
		}
		rv := rowView{
			Kind:      kind,
			Name:      l.Name,
			Quantity:  l.Quantity,
			Unit:      l.Unit,
			UnitPrice: format.Cents(l.UnitPrice),
			TaxRate:   taxRateDisplay,
			Total:     format.Cents(lineHT),
		}
		if isOption {
			optionTotalHT += lineHT
			optionRows = append(optionRows, rv)
		} else {
			totalHT += lineHT
			accumulateVAT(lineHT, taxRateStr)
			rows = append(rows, rv)
		}
	}

	for _, l := range in.Lines {
		d := parseLineData(l.Data)
		kind := d.Kind
		if kind == "" || kind == "fee" {
			kind = "line"
		}

		switch kind {
		case "group":
			rows = append(rows, rowView{Kind: "group", Name: l.Name})

		case "text":
			rows = append(rows, rowView{Kind: "text", Name: l.Name})

		case "subline":
			appendLeaf("subline", l, d.Option)

		case "detailed":
			taxRateStr := taxRateFor(l)
			detailTotal := int64(0)
			subRows := make([]rowView, 0, len(d.Sublines))
			for _, sl := range d.Sublines {
				slQty := parseQuantity(sl.Quantity)
				slHT := int64(slQty * float64(sl.UnitPrice))
				subRows = append(subRows, rowView{
					Kind:      "subline",
					Name:      sl.Name,
					Quantity:  sl.Quantity,
					Unit:      sl.Unit,
					UnitPrice: format.Cents(sl.UnitPrice),
					Total:     format.Cents(slHT),
				})
				if !sl.Option {
					detailTotal += slHT
				}
			}
			totalHT += detailTotal
			accumulateVAT(detailTotal, taxRateStr)
			rows = append(rows, rowView{
				Kind:     "detailed",
				Name:     l.Name,
				Total:    format.Cents(detailTotal),
				Sublines: subRows,
			})

		default: // "line"
			appendLeaf("line", l, d.Option)
		}
	}

	vat := make([]vatView, 0, len(rateOrder))
	for _, rate := range rateOrder {
		acc := vatByRate[rate]
		vat = append(vat, vatView{
			Rate: format.Rate(rate),
			Base: format.Cents(acc.baseHT),
			VAT:  format.Cents(acc.vatCent),
		})
	}

	optionHTStr := ""
	if optionTotalHT > 0 {
		optionHTStr = format.Cents(optionTotalHT)
	}

	return viewModel{
		ShortID:                 format.ShortID(in.Quote.QuoteId),
		QuoteName:               in.Quote.Name,
		IssuedAt:                format.Date(in.Quote.IssuedAt),
		ValidUntil:              format.Date(in.Quote.ValidUntil),
		PaymentTerms:            in.Quote.PaymentTerms,
		Sender:                  buildSender(in.User, in.UserAddress),
		Recipient:               buildRecipient(in.Client, in.ClientAddress),
		Rows:                    rows,
		OptionRows:              optionRows,
		OptionTotalHT:           optionHTStr,
		VatBreakdown:            vat,
		TotalHT:                 format.Cents(totalHT),
		TotalVAT:                format.Cents(totalVAT),
		TotalTTC:                format.Cents(totalHT + totalVAT),
		SenderSignatureLabel:    senderSignatureLabel(in.User),
		RecipientSignatureLabel: recipientSignatureLabel(in.Client),
	}
}

func buildSender(u *users.User, a *users.Address) partyView {
	v := partyView{}
	if u != nil {
		v.Title = u.Company
	}
	v.Lines = appendAddressLines(v.Lines, a)
	if u != nil {
		if u.Email != "" {
			v.Lines = append(v.Lines, u.Email)
		}
		if u.Phone != "" {
			v.Lines = append(v.Lines, u.Phone)
		}
		if u.Siret != "" {
			v.Lines = append(v.Lines, "SIRET : "+u.Siret)
		} else if u.Siren != "" {
			v.Lines = append(v.Lines, "SIREN : "+u.Siren)
		}
		if u.Vat != "" {
			v.Lines = append(v.Lines, "TVA : "+u.Vat)
		}
	}
	return v
}

func buildRecipient(c *users.Client, a *users.Address) partyView {
	v := partyView{}
	if c != nil {
		v.Title = strings.TrimSpace(c.FirstName + " " + c.LastName)
		if c.Company != "" {
			v.Lines = append(v.Lines, c.Company)
		}
	}
	v.Lines = appendAddressLines(v.Lines, a)
	if c != nil && c.Email != "" {
		v.Lines = append(v.Lines, c.Email)
	}
	return v
}

func appendAddressLines(dst []string, a *users.Address) []string {
	if a == nil {
		return dst
	}
	if a.Street != "" {
		dst = append(dst, a.Street)
	}
	if a.AdditionalStreet != "" {
		dst = append(dst, a.AdditionalStreet)
	}
	if cityLine := strings.TrimSpace(a.ZipCode + " " + a.City); cityLine != "" {
		dst = append(dst, cityLine)
	}
	return dst
}

func senderSignatureLabel(u *users.User) string {
	if u != nil && u.Company != "" {
		return "Signature " + u.Company
	}
	return "Signature de l'émetteur"
}

func recipientSignatureLabel(c *users.Client) string {
	if c != nil {
		if name := strings.TrimSpace(c.FirstName + " " + c.LastName); name != "" {
			return "Signature " + name
		}
		if c.Company != "" {
			return "Signature " + c.Company
		}
	}
	return "Signature du client"
}

func parseQuantity(s string) float64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64)
	if err != nil {
		return 0
	}
	return v
}
