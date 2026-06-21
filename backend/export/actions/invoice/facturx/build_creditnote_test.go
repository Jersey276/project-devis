package facturx

import (
	"encoding/xml"
	"strings"
	"testing"

	invoicepb "project-devis-export/services/invoice"
)

// sampleCreditNote is a single-line issued credit note referencing invoice
// 2026-0001. Amounts are stored positive (the snapshot convention).
func sampleCreditNote() *invoicepb.CreditNoteDetails {
	return &invoicepb.CreditNoteDetails{
		CreditNoteNumber: "AV-2026-0001",
		InvoiceNumber:    "2026-0001",
		IssuedAt:         "2026-06-15T09:00:00Z",
		Issuer: &invoicepb.InvoiceParty{
			Company: "Acme SARL", Siren: "123456782", Vat: "FR12345678901",
			Street: "1 rue A", ZipCode: "75001", City: "Paris",
		},
		Client: &invoicepb.InvoiceParty{
			Company: "Buyer SAS", Siren: "987654321",
			Street: "2 rue B", ZipCode: "69001", City: "Lyon",
		},
		Lines: []*invoicepb.InvoiceLine{
			{Name: "Presta A", Quantity: "1", UnitPriceCents: 10000, LineHtCents: 10000, TaxRate: "20"},
		},
		VatBreakdown: []*invoicepb.InvoiceVatLine{
			{TaxRate: "20", BaseHtCents: 10000, VatCents: 2000},
		},
		TotalHtCents:  10000,
		TotalVatCents: 2000,
		TotalTtcCents: 12000,
	}
}

func mustBuildCreditNote(t *testing.T, cn *invoicepb.CreditNoteDetails) string {
	t.Helper()
	out, err := BuildCreditNote(cn)
	if err != nil {
		t.Fatalf("BuildCreditNote: %v", err)
	}
	if err := xml.Unmarshal(out, new(struct{ XMLName xml.Name })); err != nil {
		t.Fatalf("generated XML is not well-formed: %v", err)
	}
	return string(out)
}

func TestBuildCreditNote_TypeCodeAndReference(t *testing.T) {
	s := mustBuildCreditNote(t, sampleCreditNote())
	// 381 = credit note (not 380).
	mustContain(t, s, `<ram:TypeCode>381</ram:TypeCode>`)
	mustContain(t, s, `<ram:ID>AV-2026-0001</ram:ID>`)
	// BT-3: the referenced original invoice.
	mustContain(t, s, `<ram:InvoiceReferencedDocument>`)
	mustContain(t, s, `<ram:IssuerAssignedID>2026-0001</ram:IssuerAssignedID>`)
}

func TestBuildCreditNote_PositiveAmounts(t *testing.T) {
	s := mustBuildCreditNote(t, sampleCreditNote())
	// CII credit notes carry positive amounts; the type code signals the credit.
	mustContain(t, s, `<ram:GrandTotalAmount>120.00</ram:GrandTotalAmount>`)
	mustContain(t, s, `<ram:DuePayableAmount>120.00</ram:DuePayableAmount>`)
	if strings.Contains(s, "-") && strings.Contains(s, "<ram:GrandTotalAmount>-") {
		t.Error("credit note amounts must be positive in CII")
	}
}

func TestBuildCreditNote_RejectsMissingInvoiceRef(t *testing.T) {
	cn := sampleCreditNote()
	cn.InvoiceNumber = ""
	if _, err := BuildCreditNote(cn); err == nil || !strings.Contains(err.Error(), "BT-3") {
		t.Fatalf("BuildCreditNote should reject a missing invoice reference, got err=%v", err)
	}
}

func TestBuildCreditNote_RejectsMissingNumber(t *testing.T) {
	cn := sampleCreditNote()
	cn.CreditNoteNumber = ""
	if _, err := BuildCreditNote(cn); err == nil || !strings.Contains(err.Error(), "no number") {
		t.Fatalf("BuildCreditNote should reject an unnumbered credit note, got err=%v", err)
	}
}

// A credit note inherits the seller's IBAN/BIC from the source invoice, so it
// carries the same BG-16 payment instructions.
func TestBuildCreditNote_PaymentMeansInherited(t *testing.T) {
	cn := sampleCreditNote()
	cn.Issuer.Iban = "FR7630006000011234567890189"
	cn.Issuer.Bic = "BNPAFRPP"
	s := mustBuildCreditNote(t, cn)
	mustContain(t, s, `<ram:TypeCode>30</ram:TypeCode>`)
	mustContain(t, s, `<ram:IBANID>FR7630006000011234567890189</ram:IBANID>`)
	mustContain(t, s, `<ram:BICID>BNPAFRPP</ram:BICID>`)
}

// A credit note has no due date; the settlement must not emit payment terms.
func TestBuildCreditNote_NoPaymentTerms(t *testing.T) {
	s := mustBuildCreditNote(t, sampleCreditNote())
	if strings.Contains(s, "<ram:SpecifiedTradePaymentTerms>") {
		t.Error("credit note must not carry payment terms (no due date)")
	}
}
