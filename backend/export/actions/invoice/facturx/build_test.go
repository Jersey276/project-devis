package facturx

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"

	invoicepb "project-devis-export/services/invoice"
)

// sampleInvoice is a two-line, two-rate issued invoice used across the tests.
func sampleInvoice() *invoicepb.InvoiceDetails {
	return &invoicepb.InvoiceDetails{
		InvoiceNumber: "2026-0001",
		IssuedAt:      "2026-06-14T10:30:00Z",
		SaleDate:      "2026-06-10",
		DueDate:       "2026-07-14",
		Issuer: &invoicepb.InvoiceParty{
			Company: "Acme SARL", Siren: "123456782", Vat: "FR12345678901",
			Street: "1 rue A", ZipCode: "75001", City: "Paris",
		},
		Client: &invoicepb.InvoiceParty{
			Company: "Buyer SAS", Siren: "987654321", Vat: "FR99887766554",
			Street: "2 rue B", ZipCode: "69001", City: "Lyon",
		},
		Lines: []*invoicepb.InvoiceLine{
			{Name: "Presta A", Quantity: "1", UnitPriceCents: 10000, LineHtCents: 10000, TaxRate: "20"},
			{Name: "Presta B", Quantity: "2", UnitPriceCents: 2500, LineHtCents: 5000, TaxRate: "10"},
		},
		VatBreakdown: []*invoicepb.InvoiceVatLine{
			{TaxRate: "20", BaseHtCents: 10000, VatCents: 2000},
			{TaxRate: "10", BaseHtCents: 5000, VatCents: 500},
		},
		TotalHtCents:  15000,
		TotalVatCents: 2500,
		TotalTtcCents: 17500,
	}
}

// exemptInvoice is a French franchise (art. 293 B) invoice: the seller is not a
// taxable person, so it holds no VAT number and the operation is "not subject to
// VAT" (category O). Used by both the build and the Schematron conformance tests.
func exemptInvoice() *invoicepb.InvoiceDetails {
	in := sampleInvoice()
	in.VatExempt = true
	in.VatBreakdown = nil
	in.TotalVatCents = 0
	in.TotalTtcCents = in.TotalHtCents
	// A franchise seller (and typically its buyer) carries no VAT number.
	in.Issuer.Vat = ""
	in.Client.Vat = ""
	return in
}

// Go's encoding/xml does not round-trip prefixed namespaces through XMLName, so
// we assert on the generated wire text (which a real CII consumer resolves by
// namespace URI). We do verify the document is at least well-formed XML.

func mustBuild(t *testing.T, in *invoicepb.InvoiceDetails) string {
	t.Helper()
	out, err := Build(in)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if err := xml.Unmarshal(out, new(struct{ XMLName xml.Name })); err != nil {
		t.Fatalf("generated XML is not well-formed: %v", err)
	}
	return string(out)
}

// count returns the number of occurrences of sub in s.
func count(s, sub string) int { return strings.Count(s, sub) }

func mustContain(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Errorf("generated XML missing %q", sub)
	}
}

func TestBuild_HeaderAndProfile(t *testing.T) {
	out, err := Build(sampleInvoice())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !bytes.HasPrefix(out, []byte(`<?xml version="1.0" encoding="UTF-8"?>`)) {
		t.Error("missing XML declaration")
	}
	s := string(out)
	mustContain(t, s, `xmlns:rsm="`+nsRSM+`"`)
	mustContain(t, s, `<ram:ID>`+guidelineEN16931+`</ram:ID>`)
	mustContain(t, s, `<ram:ID>2026-0001</ram:ID>`)
	mustContain(t, s, `<ram:TypeCode>380</ram:TypeCode>`)
	mustContain(t, s, `<udt:DateTimeString format="102">20260614</udt:DateTimeString>`)
}

func TestBuild_Parties(t *testing.T) {
	s := mustBuild(t, sampleInvoice())
	mustContain(t, s, `<ram:Name>Acme SARL</ram:Name>`)
	mustContain(t, s, `<ram:ID schemeID="0002">123456782</ram:ID>`) // seller SIREN
	mustContain(t, s, `<ram:ID schemeID="VA">FR12345678901</ram:ID>`)
	mustContain(t, s, `<ram:Name>Buyer SAS</ram:Name>`)
	mustContain(t, s, `<ram:ID schemeID="0002">987654321</ram:ID>`) // buyer SIREN
	mustContain(t, s, `<ram:CountryID>FR</ram:CountryID>`)
}

func TestBuild_TaxesAndTotals(t *testing.T) {
	s := mustBuild(t, sampleInvoice())
	// One header ApplicableTradeTax per rate. Header taxes carry BasisAmount;
	// line taxes do not — so counting BasisAmount yields exactly the 2 rates.
	if n := count(s, "<ram:BasisAmount>"); n != 2 {
		t.Errorf("header tax groups (BasisAmount) = %d; want 2", n)
	}
	mustContain(t, s, `<ram:TaxBasisTotalAmount>150.00</ram:TaxBasisTotalAmount>`)
	mustContain(t, s, `<ram:TaxTotalAmount currencyID="EUR">25.00</ram:TaxTotalAmount>`)
	mustContain(t, s, `<ram:GrandTotalAmount>175.00</ram:GrandTotalAmount>`)
	mustContain(t, s, `<ram:DuePayableAmount>175.00</ram:DuePayableAmount>`)
	mustContain(t, s, `<ram:BasisAmount>100.00</ram:BasisAmount>`)
}

func TestBuild_Lines(t *testing.T) {
	s := mustBuild(t, sampleInvoice())
	if n := count(s, "<ram:IncludedSupplyChainTradeLineItem>"); n != 2 {
		t.Errorf("line item count = %d; want 2", n)
	}
	mustContain(t, s, `<ram:LineID>1</ram:LineID>`)
	mustContain(t, s, `<ram:LineID>2</ram:LineID>`)
	mustContain(t, s, `<ram:BilledQuantity unitCode="C62">1</ram:BilledQuantity>`)
	mustContain(t, s, `<ram:LineTotalAmount>100.00</ram:LineTotalAmount>`)
	// Line taxes must NOT emit empty amount elements.
	if strings.Contains(s, "<ram:CalculatedAmount></ram:CalculatedAmount>") {
		t.Error("line tax emitted an empty CalculatedAmount element")
	}
}

func TestBuild_VatExempt(t *testing.T) {
	in := exemptInvoice()
	s := mustBuild(t, in)
	// Franchise art. 293 B = not subject to VAT = category O (not E "Exempt").
	mustContain(t, s, `<ram:CategoryCode>O</ram:CategoryCode>`)
	mustContain(t, s, `<ram:ExemptionReason>`+exemptReason293B+`</ram:ExemptionReason>`)
	if strings.Contains(s, "<ram:RateApplicablePercent>") {
		t.Error("a not-subject-to-VAT invoice must not carry a VAT rate")
	}
	// BR-O-02/03/04: a not-subject-to-VAT document carries no VAT number anywhere.
	if strings.Contains(s, `schemeID="VA"`) {
		t.Error("franchise (category O) invoice must not emit any VAT registration")
	}
}

func TestBuild_ClientWithoutTaxIds(t *testing.T) {
	in := sampleInvoice()
	in.Client = &invoicepb.InvoiceParty{FirstName: "Jean", LastName: "Dupont", ZipCode: "75002", City: "Paris"}
	s := mustBuild(t, in)
	mustContain(t, s, `<ram:Name>Jean Dupont</ram:Name>`)
	// Buyer block must not reference SIREN/VAT it does not have.
	if count(s, `schemeID="0002"`) != 1 { // only the seller's SIREN remains
		t.Errorf("buyer without SIREN should leave a single (seller) SIREN, got %d", count(s, `schemeID="0002"`))
	}
}

// ossInvoice is an OSS distance-selling invoice: a B2C German buyer taxed at the
// German destination rate (19%), category S. The seller stays FR.
func ossInvoice() *invoicepb.InvoiceDetails {
	return &invoicepb.InvoiceDetails{
		InvoiceNumber: "2026-0002",
		IssuedAt:      "2026-06-14T10:30:00Z",
		SaleDate:      "2026-06-10",
		DueDate:       "2026-07-14",
		OssApplied:    true,
		Issuer: &invoicepb.InvoiceParty{
			Company: "Acme SARL", Siren: "123456782", Vat: "FR12345678901",
			Street: "1 rue A", ZipCode: "75001", City: "Paris", CountryCode: "FR",
		},
		Client: &invoicepb.InvoiceParty{
			FirstName: "Hans", LastName: "Müller",
			Street: "Hauptstr. 3", ZipCode: "10115", City: "Berlin", CountryCode: "DE",
		},
		Lines: []*invoicepb.InvoiceLine{
			{Name: "Presta A", Quantity: "1", UnitPriceCents: 10000, LineHtCents: 10000, TaxRate: "19"},
		},
		VatBreakdown: []*invoicepb.InvoiceVatLine{
			{TaxRate: "19", BaseHtCents: 10000, VatCents: 1900},
		},
		TotalHtCents:  10000,
		TotalVatCents: 1900,
		TotalTtcCents: 11900,
	}
}

// TestBuild_OSS verifies an OSS invoice emits the real buyer country (DE) and
// the destination rate under the standard category S — never the FR default.
func TestBuild_OSS(t *testing.T) {
	s := mustBuild(t, ossInvoice())
	mustContain(t, s, `<ram:CountryID>DE</ram:CountryID>`) // buyer in Germany
	mustContain(t, s, `<ram:CountryID>FR</ram:CountryID>`) // seller stays FR
	mustContain(t, s, `<ram:CategoryCode>S</ram:CategoryCode>`)
	mustContain(t, s, `<ram:RateApplicablePercent>19.00</ram:RateApplicablePercent>`)
	// No buyer in FR: the only FR country must be the seller's.
	if n := count(s, `<ram:CountryID>FR</ram:CountryID>`); n != 1 {
		t.Errorf("FR country count = %d; want 1 (seller only)", n)
	}
}

func TestBuild_PaymentMeans(t *testing.T) {
	in := sampleInvoice()
	in.Issuer.Iban = "FR7630006000011234567890189"
	in.Issuer.Bic = "BNPAFRPP"
	s := mustBuild(t, in)

	mustContain(t, s, `<ram:SpecifiedTradeSettlementPaymentMeans>`)
	mustContain(t, s, `<ram:TypeCode>30</ram:TypeCode>`) // BT-81 credit transfer
	mustContain(t, s, `<ram:IBANID>FR7630006000011234567890189</ram:IBANID>`) // BT-84
	mustContain(t, s, `<ram:BICID>BNPAFRPP</ram:BICID>`)                      // BT-86

	// PaymentMeans must sit after InvoiceCurrencyCode and before the header tax;
	// anchor on the currency code to skip the earlier line-level taxes.
	cur := strings.Index(s, "InvoiceCurrencyCode")
	pm := strings.Index(s, "SpecifiedTradeSettlementPaymentMeans")
	headerTax := strings.Index(s[cur:], "ApplicableTradeTax") + cur
	if pm == -1 || pm < cur || pm > headerTax {
		t.Errorf("PaymentMeans (%d) must sit between InvoiceCurrencyCode (%d) and the header ApplicableTradeTax (%d)", pm, cur, headerTax)
	}
}

func TestBuild_PaymentMeans_NoBic(t *testing.T) {
	in := sampleInvoice()
	in.Issuer.Iban = "FR7630006000011234567890189"
	s := mustBuild(t, in)
	mustContain(t, s, `<ram:IBANID>FR7630006000011234567890189</ram:IBANID>`)
	// BIC is optional: no empty institution element when absent.
	if strings.Contains(s, "PayeeSpecifiedCreditorFinancialInstitution") {
		t.Error("missing BIC must not emit a PayeeSpecifiedCreditorFinancialInstitution element")
	}
}

// Without an issuer IBAN the conditional BG-16 group is omitted entirely, keeping
// franchise / B2C invoices EN 16931-valid.
func TestBuild_NoPaymentMeansWithoutIban(t *testing.T) {
	s := mustBuild(t, sampleInvoice())
	if strings.Contains(s, "SpecifiedTradeSettlementPaymentMeans") {
		t.Error("no IBAN should omit the whole PaymentMeans group")
	}
}

func TestBuild_RejectsDraft(t *testing.T) {
	in := sampleInvoice()
	in.InvoiceNumber = ""
	if _, err := Build(in); err == nil || !strings.Contains(err.Error(), "no number") {
		t.Fatalf("Build should reject an unnumbered (draft) invoice, got err=%v", err)
	}
}

func TestBuild_RejectsInconsistentTotals(t *testing.T) {
	in := sampleInvoice()
	in.TotalHtCents = 99999
	if _, err := Build(in); err == nil || !strings.Contains(err.Error(), "HT") {
		t.Fatalf("Build should reject inconsistent HT totals, got err=%v", err)
	}
}
