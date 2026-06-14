package facturx

import (
	"encoding/xml"
	"fmt"
	"strings"

	invoicepb "project-devis-export/services/invoice"
)

const xmlHeader = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"

// Build renders the EN 16931 CII XML for an issued invoice. It returns an error
// for inputs that cannot yield a valid invoice (a draft has no number; figures
// must be internally consistent), so the caller never emits a broken Factur-X.
func Build(in *invoicepb.InvoiceDetails) ([]byte, error) {
	if in == nil {
		return nil, fmt.Errorf("facturx: nil invoice")
	}
	number := strings.TrimSpace(in.GetInvoiceNumber())
	if number == "" {
		return nil, fmt.Errorf("facturx: invoice has no number (not issued); cannot build EN 16931 XML")
	}
	if len(in.GetLines()) == 0 {
		return nil, fmt.Errorf("facturx: invoice %s has no lines", number)
	}
	if err := checkConsistency(in); err != nil {
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
			ID:            number,
			TypeCode:      typeCodeInvoice,
			IssueDateTime: dateWrap(in.GetIssuedAt()),
		},
		Transaction: supplyChainTradeTransaction{
			Lines:      buildLines(in),
			Agreement:  buildAgreement(in),
			Delivery:   buildDelivery(in),
			Settlement: buildSettlement(in),
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
func checkConsistency(in *invoicepb.InvoiceDetails) error {
	var lineSum int64
	for _, l := range in.GetLines() {
		lineSum += l.GetLineHtCents()
	}
	if lineSum != in.GetTotalHtCents() {
		return fmt.Errorf("facturx: line HT sum %d != total HT %d", lineSum, in.GetTotalHtCents())
	}
	var vatSum int64
	for _, v := range in.GetVatBreakdown() {
		vatSum += v.GetVatCents()
	}
	if vatSum != in.GetTotalVatCents() {
		return fmt.Errorf("facturx: VAT breakdown sum %d != total VAT %d", vatSum, in.GetTotalVatCents())
	}
	return nil
}

func buildLines(in *invoicepb.InvoiceDetails) []lineItem {
	exempt := in.GetVatExempt()
	out := make([]lineItem, 0, len(in.GetLines()))
	for i, l := range in.GetLines() {
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

func buildAgreement(in *invoicepb.InvoiceDetails) headerTradeAgreement {
	return headerTradeAgreement{
		Seller: party(in.GetIssuer()),
		Buyer:  party(in.GetClient()),
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
		CountryID:    countryFR,
	}
}

func buildDelivery(in *invoicepb.InvoiceDetails) headerTradeDelivery {
	d := dateCII(in.GetSaleDate())
	if d == "" {
		return headerTradeDelivery{}
	}
	return headerTradeDelivery{
		Event: &supplyChainEvent{OccurrenceDateTime: dateWrapRaw(d)},
	}
}

func buildSettlement(in *invoicepb.InvoiceDetails) headerTradeSettlement {
	s := headerTradeSettlement{
		CurrencyCode: currencyEUR,
		Taxes:        buildTaxes(in),
		Summation: monetarySummation{
			LineTotalAmount:     amountFromCents(in.GetTotalHtCents()),
			TaxBasisTotalAmount: amountFromCents(in.GetTotalHtCents()),
			TaxTotalAmount:      currencyAmount{CurrencyID: currencyEUR, Value: amountFromCents(in.GetTotalVatCents())},
			GrandTotalAmount:    amountFromCents(in.GetTotalTtcCents()),
			DuePayableAmount:    amountFromCents(in.GetTotalTtcCents()),
		},
	}
	if due := dateCII(in.GetDueDate()); due != "" {
		s.PaymentTerms = &paymentTerms{DueDate: dateWrapRaw(due)}
	}
	return s
}

// buildTaxes emits one ApplicableTradeTax per VAT-breakdown bucket. Under the
// franchise the whole invoice is a single exempt group.
func buildTaxes(in *invoicepb.InvoiceDetails) []tradeTax {
	exempt := in.GetVatExempt()
	if exempt {
		return []tradeTax{{
			CalculatedAmount: amountFromCents(0),
			TypeCode:         "VAT",
			ExemptionReason:  exemptReason293B,
			BasisAmount:      amountFromCents(in.GetTotalHtCents()),
			CategoryCode:     categoryExempt,
		}}
	}
	out := make([]tradeTax, 0, len(in.GetVatBreakdown()))
	for _, v := range in.GetVatBreakdown() {
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
