// Package facturx builds the EN 16931 (Cross Industry Invoice, CII) XML payload
// of a Factur-X document from an immutable snapshot — an invoice (Build, type
// 380) or a credit note (BuildCreditNote, type 381). It is pure: it only reads
// the snapshot proto and returns XML bytes — no I/O, no PDF.
//
// Scope note: this package produces the structured XML only. Embedding it into a
// PDF/A-3 (attachment + AFRelationship + urn:factur-x XMP) is a separate step.
package facturx

import (
	"fmt"
	"strconv"
	"strings"
)

// Fixed values for the targeted profile and the French, euro context.
const (
	// guidelineEN16931 is the profile identifier (BT-24) carried by the document
	// context. The Factur-X EN 16931 "compliant" URN both names the EN 16931 model
	// and declares Factur-X conformance — what a French PDP expects.
	guidelineEN16931 = "urn:cen.eu:en16931:2017#compliant#urn:factur-x.eu:1p0:en16931"
	// typeCodeInvoice is UNTDID 1001 code 380 = commercial invoice.
	typeCodeInvoice = "380"
	// typeCodeCreditNote is UNTDID 1001 code 381 = credit note.
	typeCodeCreditNote = "381"
	currencyEUR        = "EUR"
	// countryFR: the snapshot has no structured country, FR is assumed (documented
	// limitation — revisit if cross-border issuing is ever supported).
	countryFR = "FR"
	// schemeSIREN is ISO 6523 code 0002 (French SIRENE registry).
	schemeSIREN = "0002"
	// schemeVAT is the UNTDID 1153 "VA" qualifier for a VAT registration number.
	schemeVAT = "VA"
	// unitDefault is UN/ECE Rec 20 "C62" = one (dimensionless unit).
	unitDefault = "C62"

	// paymentMeansCreditTransfer is UNTDID 4461 code 30 = credit transfer (BT-81),
	// the only payment means modelled.
	paymentMeansCreditTransfer = "30"

	// exemptReason293B is the statutory mention for the French VAT franchise.
	exemptReason293B = "TVA non applicable, art. 293 B du CGI"
)

// VAT category codes (UNTDID 5305) used by ApplicableTradeTax.
const (
	categoryStandard   = "S" // standard rate
	categoryNotSubject = "O" // not subject to VAT (art. 293 B franchise)
	categoryZero       = "Z" // zero-rated
)

// categoryForRate maps a snapshot tax rate to its EN 16931 VAT category code.
// Under the French franchise (vatExempt) the seller is NOT a taxable person — it
// holds no VAT number — so the operation is category O "Not subject to VAT", not
// E "Exempt". EN 16931 keeps these distinct: an E document requires a seller VAT
// number and an explicit 0% rate (BR-E-02/05), whereas O forbids any VAT number
// on any party and carries no rate (BR-O-02/05). The franchise is squarely O.
func categoryForRate(rate string, vatExempt bool) string {
	if vatExempt {
		return categoryNotSubject
	}
	if isZeroRate(rate) {
		return categoryZero
	}
	return categoryStandard
}

func isZeroRate(rate string) bool {
	f, err := parseRate(rate)
	return err == nil && f == 0
}

// amountFromCents renders integer cents as a fixed 2-decimal string with a dot
// separator (e.g. -12345 -> "-123.45"). EN 16931 amounts must never carry a
// currency symbol or a comma — do NOT reuse the display formatter from render.go.
func amountFromCents(cents int64) string {
	neg := cents < 0
	if neg {
		cents = -cents
	}
	s := fmt.Sprintf("%d.%02d", cents/100, cents%100)
	if neg {
		return "-" + s
	}
	return s
}

// parseRate parses a snapshot rate string ("20", "5.5", "") into a float.
func parseRate(rate string) (float64, error) {
	rate = strings.TrimSpace(rate)
	if rate == "" {
		return 0, nil
	}
	return strconv.ParseFloat(strings.ReplaceAll(rate, ",", "."), 64)
}

// percentFromRate normalises a rate string to two decimals ("20" -> "20.00").
// An empty/invalid rate yields "0.00".
func percentFromRate(rate string) string {
	f, err := parseRate(rate)
	if err != nil {
		return "0.00"
	}
	return strconv.FormatFloat(f, 'f', 2, 64)
}

// dateCII converts an RFC3339 timestamp (or a bare YYYY-MM-DD) to the CII
// "format 102" date string YYYYMMDD. Empty input yields "".
func dateCII(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	// Keep the date part before any 'T', then strip dashes.
	if i := strings.IndexByte(value, 'T'); i > 0 {
		value = value[:i]
	}
	return strings.ReplaceAll(value, "-", "")
}
