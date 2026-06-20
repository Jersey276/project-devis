package facturx

import (
	"encoding/xml"
	"fmt"
	"strings"

	invoicepb "project-devis-export/services/invoice"
)

const xmlHeader = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"

// docInput is the normalized view the CII builder works on. Both an invoice
// (380) and a credit note (381) map onto it, so the structured-document logic
// (lines, parties, taxes, totals) is written once. Fields a given document type
// lacks (a credit note has no due/sale date; an invoice has no referenced
// document) are simply left zero.
type docInput struct {
	typeCode    string
	number      string
	issuedAt    string
	saleDate    string
	dueDate     string
	referencedInvoice string // BT-3, credit note only

	issuer *invoicepb.InvoiceParty
	client *invoicepb.InvoiceParty
	lines  []*invoicepb.InvoiceLine
	vat    []*invoicepb.InvoiceVatLine

	totalHtCents  int64
	totalVatCents int64
	totalTtcCents int64
	vatExempt     bool
}

// Build renders the EN 16931 CII XML for an issued invoice. It returns an error
// for inputs that cannot yield a valid invoice (a draft has no number; figures
// must be internally consistent), so the caller never emits a broken Factur-X.
func Build(in *invoicepb.InvoiceDetails) ([]byte, error) {
	if in == nil {
		return nil, fmt.Errorf("facturx: nil invoice")
	}
	return build(docInput{
		typeCode:      typeCodeInvoice,
		number:        strings.TrimSpace(in.GetInvoiceNumber()),
		issuedAt:      in.GetIssuedAt(),
		saleDate:      in.GetSaleDate(),
		dueDate:       in.GetDueDate(),
		issuer:        in.GetIssuer(),
		client:        in.GetClient(),
		lines:         in.GetLines(),
		vat:           in.GetVatBreakdown(),
		totalHtCents:  in.GetTotalHtCents(),
		totalVatCents: in.GetTotalVatCents(),
		totalTtcCents: in.GetTotalTtcCents(),
		vatExempt:     in.GetVatExempt(),
	}, "invoice")
}

// BuildCreditNote renders the EN 16931 CII XML for an issued credit note (type
// 381). Snapshot amounts are stored positive and stay positive here: in CII the
// document type, not the sign, signals a credit. The original invoice number is
// carried as BT-3 (InvoiceReferencedDocument), mandatory for a credit note.
func BuildCreditNote(cn *invoicepb.CreditNoteDetails) ([]byte, error) {
	if cn == nil {
		return nil, fmt.Errorf("facturx: nil credit note")
	}
	ref := strings.TrimSpace(cn.GetInvoiceNumber())
	if ref == "" {
		return nil, fmt.Errorf("facturx: credit note %s references no invoice number (BT-3 required)", strings.TrimSpace(cn.GetCreditNoteNumber()))
	}
	return build(docInput{
		typeCode:          typeCodeCreditNote,
		number:            strings.TrimSpace(cn.GetCreditNoteNumber()),
		issuedAt:          cn.GetIssuedAt(),
		referencedInvoice: ref,
		issuer:            cn.GetIssuer(),
		client:            cn.GetClient(),
		lines:             cn.GetLines(),
		vat:               cn.GetVatBreakdown(),
		totalHtCents:      cn.GetTotalHtCents(),
		totalVatCents:     cn.GetTotalVatCents(),
		totalTtcCents:     cn.GetTotalTtcCents(),
		vatExempt:         cn.GetVatExempt(),
	}, "credit note")
}

// build is the shared CII document builder. kind names the document in error
// messages ("invoice" / "credit note").
func build(d docInput, kind string) ([]byte, error) {
	if d.number == "" {
		return nil, fmt.Errorf("facturx: %s has no number (not issued); cannot build EN 16931 XML", kind)
	}
	if len(d.lines) == 0 {
		return nil, fmt.Errorf("facturx: %s %s has no lines", kind, d.number)
	}
	if err := checkConsistency(d); err != nil {
		return nil, err
	}

	doc := crossIndustryInvoice{
		XMLNSrsm: nsRSM,
		XMLNSram: nsRAM,
		XMLNSudt: nsUDT,
		Context: exchangedDocumentContext{
			Guideline: guidelineParameter{ID: guidelineEN16931},
		},
		Document: exchangedDocument{
			ID:            d.number,
			TypeCode:      d.typeCode,
			IssueDateTime: dateWrap(d.issuedAt),
		},
		Transaction: supplyChainTradeTransaction{
			Lines:      buildLines(d),
			Agreement:  buildAgreement(d),
			Delivery:   buildDelivery(d),
			Settlement: buildSettlement(d),
		},
	}

	body, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("facturx: marshal: %w", err)
	}
	return append([]byte(xmlHeader), body...), nil
}

// checkConsistency guards against a snapshot whose totals do not add up — the
// generator copies frozen figures, so any mismatch is a data bug, not rounding.
func checkConsistency(d docInput) error {
	var lineSum int64
	for _, l := range d.lines {
		lineSum += l.GetLineHtCents()
	}
	if lineSum != d.totalHtCents {
		return fmt.Errorf("facturx: line HT sum %d != total HT %d", lineSum, d.totalHtCents)
	}
	var vatSum int64
	for _, v := range d.vat {
		vatSum += v.GetVatCents()
	}
	if vatSum != d.totalVatCents {
		return fmt.Errorf("facturx: VAT breakdown sum %d != total VAT %d", vatSum, d.totalVatCents)
	}
	return nil
}

func buildLines(d docInput) []lineItem {
	exempt := d.vatExempt
	out := make([]lineItem, 0, len(d.lines))
	for i, l := range d.lines {
		out = append(out, lineItem{
			DocLine: lineDocument{LineID: fmt.Sprintf("%d", i+1)},
			Product: tradeProduct{Name: l.GetName()},
			Agreement: lineTradeAgreement{
				NetPrice: tradePrice{ChargeAmount: amountFromCents(l.GetUnitPriceCents())},
			},
			Delivery: lineTradeDelivery{
				BilledQuantity: quantity{UnitCode: unitDefault, Value: quantityValue(l.GetQuantity())},
			},
			Settlement: lineTradeSettlement{
				Tax:       lineTax(l.GetTaxRate(), exempt),
				Summation: lineMonetarySummation{LineTotalAmount: amountFromCents(l.GetLineHtCents())},
			},
		})
	}
	return out
}

func lineTax(rate string, exempt bool) tradeTax {
	t := tradeTax{
		TypeCode:     "VAT",
		CategoryCode: categoryForRate(rate, exempt),
	}
	if exempt {
		t.ExemptionReason = exemptReason293B
		return t
	}
	t.RateApplicablePercent = percentFromRate(rate)
	return t
}

func buildAgreement(d docInput) headerTradeAgreement {
	return headerTradeAgreement{
		Seller: party(d.issuer),
		Buyer:  party(d.client),
	}
}

// party maps a snapshot InvoiceParty to a CII trade party. SIREN/VAT groups are
// omitted when absent (legacy or B2C rows), which stays EN 16931-valid.
func party(p *invoicepb.InvoiceParty) tradeParty {
	if p == nil {
		return tradeParty{Address: &postalAddress{CountryID: countryFR}}
	}
	tp := tradeParty{
		Name:    partyName(p),
		Address: address(p),
	}
	if siren := strings.TrimSpace(p.GetSiren()); siren != "" {
		tp.LegalOrg = &legalOrganization{ID: schemeID{SchemeID: schemeSIREN, Value: siren}}
	}
	if vat := strings.TrimSpace(p.GetVat()); vat != "" {
		tp.TaxRegistration = &taxRegistration{ID: schemeID{SchemeID: schemeVAT, Value: vat}}
	}
	return tp
}

// partyName prefers the legal company name; falls back to the individual's name
// (BT-44 buyer name is mandatory and must never be empty).
func partyName(p *invoicepb.InvoiceParty) string {
	if c := strings.TrimSpace(p.GetCompany()); c != "" {
		return c
	}
	return strings.TrimSpace(p.GetFirstName() + " " + p.GetLastName())
}

func address(p *invoicepb.InvoiceParty) *postalAddress {
	return &postalAddress{
		PostcodeCode: p.GetZipCode(),
		LineOne:      p.GetStreet(),
		LineTwo:      p.GetAdditionalStreet(),
		CityName:     p.GetCity(),
		CountryID:    countryCode(p),
	}
}

// countryCode returns the party's frozen ISO 3166-1 alpha-2 code, falling back
// to FR for legacy snapshots that predate the frozen code (preserves the prior
// hardcoded behaviour). The code is the buyer/seller country in the CII XML —
// notably the destination country for an OSS distance-selling invoice.
func countryCode(p *invoicepb.InvoiceParty) string {
	if c := strings.ToUpper(strings.TrimSpace(p.GetCountryCode())); c != "" {
		return c
	}
	return countryFR
}

func buildDelivery(d docInput) headerTradeDelivery {
	dt := dateCII(d.saleDate)
	if dt == "" {
		return headerTradeDelivery{}
	}
	return headerTradeDelivery{
		Event: &supplyChainEvent{OccurrenceDateTime: dateWrapRaw(dt)},
	}
}

func buildSettlement(d docInput) headerTradeSettlement {
	s := headerTradeSettlement{
		CurrencyCode: currencyEUR,
		Taxes:        buildTaxes(d),
		Summation: monetarySummation{
			LineTotalAmount:     amountFromCents(d.totalHtCents),
			TaxBasisTotalAmount: amountFromCents(d.totalHtCents),
			TaxTotalAmount:      currencyAmount{CurrencyID: currencyEUR, Value: amountFromCents(d.totalVatCents)},
			GrandTotalAmount:    amountFromCents(d.totalTtcCents),
			DuePayableAmount:    amountFromCents(d.totalTtcCents),
		},
	}
	if due := dateCII(d.dueDate); due != "" {
		s.PaymentTerms = &paymentTerms{DueDate: dateWrapRaw(due)}
	}
	if d.referencedInvoice != "" {
		s.InvoiceReferenced = &referencedDocument{IssuerAssignedID: d.referencedInvoice}
	}
	return s
}

// buildTaxes emits one ApplicableTradeTax per VAT-breakdown bucket. Under the
// franchise the whole document is a single exempt group.
func buildTaxes(d docInput) []tradeTax {
	if d.vatExempt {
		return []tradeTax{{
			CalculatedAmount: amountFromCents(0),
			TypeCode:         "VAT",
			ExemptionReason:  exemptReason293B,
			BasisAmount:      amountFromCents(d.totalHtCents),
			CategoryCode:     categoryExempt,
		}}
	}
	out := make([]tradeTax, 0, len(d.vat))
	for _, v := range d.vat {
		out = append(out, tradeTax{
			CalculatedAmount:      amountFromCents(v.GetVatCents()),
			TypeCode:              "VAT",
			BasisAmount:           amountFromCents(v.GetBaseHtCents()),
			CategoryCode:          categoryForRate(v.GetTaxRate(), false),
			RateApplicablePercent: percentFromRate(v.GetTaxRate()),
		})
	}
	return out
}

// dateWrap builds a format-102 DateTimeString from an RFC3339/date value.
func dateWrap(value string) dateTimeWrap {
	return dateWrapRaw(dateCII(value))
}

// dateWrapRaw wraps an already-formatted YYYYMMDD value.
func dateWrapRaw(yyyymmdd string) dateTimeWrap {
	return dateTimeWrap{DateTimeString: formattedDate{Format: "102", Value: yyyymmdd}}
}

// quantityValue defaults an empty quantity to "1" (a CII BilledQuantity is
// mandatory and must be numeric).
func quantityValue(q string) string {
	q = strings.TrimSpace(q)
	if q == "" {
		return "1"
	}
	return q
}
