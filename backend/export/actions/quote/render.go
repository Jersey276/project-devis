package quote

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"project-devis-export/quote"
	"project-devis-export/services/gotenberg"
	"project-devis-export/templates"
	"project-devis-export/users"
)

// quoteTpl is parsed once at package init. html/template is used (not
// text/template) so every {{.Foo}} interpolation is auto-escaped, which
// shields us from injection via quote names, client names, etc.
var quoteTpl = template.Must(template.New("quote.html").Parse(string(templates.QuoteHTML)))

// viewModel is the shape consumed by quote.html. It's intentionally flat — no
// proto types, no formatting helpers in the template — so the template stays
// readable and so a future templating system (premium templates) gets a clean
// contract to bind against.
type viewModel struct {
	ShortID                  string
	Quote                    quoteView
	Sender                   senderView
	Recipient                recipientView
	Lines                    []lineView
	TotalHT                  string
	SenderSignatureLabel     string
	RecipientSignatureLabel  string
}

type quoteView struct {
	Name string
}

type senderView struct {
	LogoURL string
	Company string
	Lines   []string
}

type recipientView struct {
	Name  string
	Lines []string
}

type lineView struct {
	Name      string
	Quantity  string
	Unit      string
	UnitPrice string
	Total     string
}

func Render(ctx context.Context, gt *gotenberg.Client,
	q *quote.Quote, lines []*quote.QuoteLine,
	u *users.User, ua *users.Address,
	c *users.Client, ca *users.Address) ([]byte, error) {

	vm := buildViewModel(q, lines, u, ua, c, ca)

	var html bytes.Buffer
	if err := quoteTpl.Execute(&html, vm); err != nil {
		return nil, fmt.Errorf("render template: %w", err)
	}
	return gt.Convert(ctx, html.Bytes())
}

func buildViewModel(q *quote.Quote, lines []*quote.QuoteLine,
	u *users.User, ua *users.Address,
	c *users.Client, ca *users.Address) viewModel {

	totalCents := int64(0)
	lineViews := make([]lineView, 0, len(lines))
	for _, l := range lines {
		qty := parseQuantity(l.Quantity)
		lineTotal := int64(qty * float64(l.UnitPrice))
		totalCents += lineTotal
		lineViews = append(lineViews, lineView{
			Name:      l.Name,
			Quantity:  l.Quantity,
			Unit:      l.Unit,
			UnitPrice: formatCents(l.UnitPrice),
			Total:     formatCents(lineTotal),
		})
	}

	return viewModel{
		ShortID:                 shortID(q.QuoteId),
		Quote:                   quoteView{Name: q.Name},
		Sender:                  buildSender(u, ua),
		Recipient:               buildRecipient(c, ca),
		Lines:                   lineViews,
		TotalHT:                 formatCents(totalCents),
		SenderSignatureLabel:    senderSignatureLabel(u),
		RecipientSignatureLabel: recipientSignatureLabel(c),
	}
}

func buildSender(u *users.User, a *users.Address) senderView {
	v := senderView{}
	if u != nil {
		v.LogoURL = u.LogoUrl
		v.Company = u.Company
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

func buildRecipient(c *users.Client, a *users.Address) recipientView {
	v := recipientView{}
	if c != nil {
		v.Name = strings.TrimSpace(c.FirstName + " " + c.LastName)
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

func shortID(id string) string {
	if len(id) >= 8 {
		return id[:8]
	}
	return id
}
