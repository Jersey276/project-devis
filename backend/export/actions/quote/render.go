package quote

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"project-devis-export/internal/format"
	"project-devis-export/quote"
	"project-devis-export/templates"
	"project-devis-export/users"
)

var quoteTpl = template.Must(template.New("quote.html").Parse(string(templates.QuoteHTML)))

type renderInput struct {
	Quote         *quote.Quote
	Lines         []*quote.QuoteLine
	User          *users.User
	UserAddress   *users.Address
	Client        *users.Client
	ClientAddress *users.Address
}

type viewModel struct {
	ShortID                 string
	QuoteName               string
	Sender                  partyView
	Recipient               partyView
	Lines                   []lineView
	TotalHT                 string
	SenderSignatureLabel    string
	RecipientSignatureLabel string
}

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
	Total     string
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
	totalCents := int64(0)
	lineViews := make([]lineView, 0, len(in.Lines))
	for _, l := range in.Lines {
		qty := parseQuantity(l.Quantity)
		lineTotal := int64(qty * float64(l.UnitPrice))
		totalCents += lineTotal
		lineViews = append(lineViews, lineView{
			Name:      l.Name,
			Quantity:  l.Quantity,
			Unit:      l.Unit,
			UnitPrice: format.Cents(l.UnitPrice),
			Total:     format.Cents(lineTotal),
		})
	}

	return viewModel{
		ShortID:                 format.ShortID(in.Quote.QuoteId),
		QuoteName:               in.Quote.Name,
		Sender:                  buildSender(in.User, in.UserAddress),
		Recipient:               buildRecipient(in.Client, in.ClientAddress),
		Lines:                   lineViews,
		TotalHT:                 format.Cents(totalCents),
		SenderSignatureLabel:    senderSignatureLabel(in.User),
		RecipientSignatureLabel: recipientSignatureLabel(in.Client),
	}
}

func buildSender(u *users.User, a *users.Address) partyView {
	v := partyView{}
	if u != nil {
		v.LogoURL = u.LogoUrl
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
		if u.Siren != "" {
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
