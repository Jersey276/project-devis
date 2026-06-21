package facturx

import "encoding/xml"

// CII namespaces (Factur-X / ZUGFeRD 2.x, EN 16931 profile).
const (
	nsRSM = "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"
	nsRAM = "urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100"
	nsUDT = "urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100"
)

// crossIndustryInvoice is the document root. Namespaces are declared once here;
// every child element is prefixed via its xml tag.
type crossIndustryInvoice struct {
	XMLName xml.Name `xml:"rsm:CrossIndustryInvoice"`
	XMLNSrsm string  `xml:"xmlns:rsm,attr"`
	XMLNSram string  `xml:"xmlns:ram,attr"`
	XMLNSudt string  `xml:"xmlns:udt,attr"`

	Context     exchangedDocumentContext   `xml:"rsm:ExchangedDocumentContext"`
	Document    exchangedDocument          `xml:"rsm:ExchangedDocument"`
	Transaction supplyChainTradeTransaction `xml:"rsm:SupplyChainTradeTransaction"`
}

type exchangedDocumentContext struct {
	Guideline guidelineParameter `xml:"ram:GuidelineSpecifiedDocumentContextParameter"`
}

type guidelineParameter struct {
	ID string `xml:"ram:ID"`
}

type exchangedDocument struct {
	ID          string       `xml:"ram:ID"`
	TypeCode    string       `xml:"ram:TypeCode"`
	IssueDateTime dateTimeWrap `xml:"ram:IssueDateTime"`
}

// dateTimeWrap carries a CII DateTimeString with its format qualifier (102 =
// YYYYMMDD). The wrapping element name is set by the parent's xml tag.
type dateTimeWrap struct {
	DateTimeString formattedDate `xml:"udt:DateTimeString"`
}

type formattedDate struct {
	Format string `xml:"format,attr"`
	Value  string `xml:",chardata"`
}

type supplyChainTradeTransaction struct {
	Lines      []lineItem               `xml:"ram:IncludedSupplyChainTradeLineItem"`
	Agreement  headerTradeAgreement     `xml:"ram:ApplicableHeaderTradeAgreement"`
	Delivery   headerTradeDelivery      `xml:"ram:ApplicableHeaderTradeDelivery"`
	Settlement headerTradeSettlement    `xml:"ram:ApplicableHeaderTradeSettlement"`
}

// ─── Line items ──────────────────────────────────────────────────────────────

type lineItem struct {
	DocLine    lineDocument         `xml:"ram:AssociatedDocumentLineDocument"`
	Product    tradeProduct         `xml:"ram:SpecifiedTradeProduct"`
	Agreement  lineTradeAgreement   `xml:"ram:SpecifiedLineTradeAgreement"`
	Delivery   lineTradeDelivery    `xml:"ram:SpecifiedLineTradeDelivery"`
	Settlement lineTradeSettlement  `xml:"ram:SpecifiedLineTradeSettlement"`
}

type lineDocument struct {
	LineID string `xml:"ram:LineID"`
}

type tradeProduct struct {
	Name string `xml:"ram:Name"`
}

type lineTradeAgreement struct {
	NetPrice tradePrice `xml:"ram:NetPriceProductTradePrice"`
}

type tradePrice struct {
	ChargeAmount string `xml:"ram:ChargeAmount"`
}

type lineTradeDelivery struct {
	BilledQuantity quantity `xml:"ram:BilledQuantity"`
}

type quantity struct {
	UnitCode string `xml:"unitCode,attr"`
	Value    string `xml:",chardata"`
}

type lineTradeSettlement struct {
	Tax        tradeTax              `xml:"ram:ApplicableTradeTax"`
	Summation  lineMonetarySummation `xml:"ram:SpecifiedTradeSettlementLineMonetarySummation"`
}

type lineMonetarySummation struct {
	LineTotalAmount string `xml:"ram:LineTotalAmount"`
}

// ─── Header: agreement (parties) ─────────────────────────────────────────────

type headerTradeAgreement struct {
	Seller tradeParty `xml:"ram:SellerTradeParty"`
	Buyer  tradeParty `xml:"ram:BuyerTradeParty"`
}

type tradeParty struct {
	Name            string             `xml:"ram:Name"`
	LegalOrg        *legalOrganization `xml:"ram:SpecifiedLegalOrganization,omitempty"`
	Address         *postalAddress     `xml:"ram:PostalTradeAddress,omitempty"`
	TaxRegistration *taxRegistration   `xml:"ram:SpecifiedTaxRegistration,omitempty"`
}

type legalOrganization struct {
	ID schemeID `xml:"ram:ID"`
}

type schemeID struct {
	SchemeID string `xml:"schemeID,attr"`
	Value    string `xml:",chardata"`
}

type postalAddress struct {
	PostcodeCode string `xml:"ram:PostcodeCode,omitempty"`
	LineOne      string `xml:"ram:LineOne,omitempty"`
	LineTwo      string `xml:"ram:LineTwo,omitempty"`
	CityName     string `xml:"ram:CityName,omitempty"`
	CountryID    string `xml:"ram:CountryID"`
}

type taxRegistration struct {
	ID schemeID `xml:"ram:ID"`
}

// ─── Header: delivery ────────────────────────────────────────────────────────

type headerTradeDelivery struct {
	Event *supplyChainEvent `xml:"ram:ActualDeliverySupplyChainEvent,omitempty"`
}

type supplyChainEvent struct {
	OccurrenceDateTime dateTimeWrap `xml:"ram:OccurrenceDateTime"`
}

// ─── Header: settlement (taxes, payment, totals) ─────────────────────────────

type headerTradeSettlement struct {
	CurrencyCode string             `xml:"ram:InvoiceCurrencyCode"`
	// BG-16: must precede ApplicableTradeTax per the CII XSD sequence (field order
	// is the marshalled order).
	PaymentMeans []paymentMeans     `xml:"ram:SpecifiedTradeSettlementPaymentMeans,omitempty"`
	Taxes        []tradeTax         `xml:"ram:ApplicableTradeTax"`
	PaymentTerms *paymentTerms      `xml:"ram:SpecifiedTradePaymentTerms,omitempty"`
	Summation    monetarySummation  `xml:"ram:SpecifiedTradeSettlementHeaderMonetarySummation"`
	// InvoiceReferenced carries BT-3 — the original invoice a credit note (381)
	// rebills against. Absent on a plain invoice (380).
	InvoiceReferenced *referencedDocument `xml:"ram:InvoiceReferencedDocument,omitempty"`
}

// referencedDocument is a CII ReferencedDocument: the credit note points back at
// its source invoice by number (BT-3). The optional FormattedIssueDateTime is
// omitted — it lives in the qdt namespace and the snapshot does not carry the
// original invoice's issue date, while BT-3 only requires the number.
type referencedDocument struct {
	IssuerAssignedID string `xml:"ram:IssuerAssignedID"`
}

// tradeTax is used both at header level (CalculatedAmount/BasisAmount filled)
// and at line level (only TypeCode/CategoryCode/RateApplicablePercent) — the
// amount fields are omitempty so a line tax does not emit empty elements.
type tradeTax struct {
	CalculatedAmount      string `xml:"ram:CalculatedAmount,omitempty"`
	TypeCode              string `xml:"ram:TypeCode"`
	ExemptionReason       string `xml:"ram:ExemptionReason,omitempty"`
	BasisAmount           string `xml:"ram:BasisAmount,omitempty"`
	CategoryCode          string `xml:"ram:CategoryCode"`
	RateApplicablePercent string `xml:"ram:RateApplicablePercent,omitempty"`
}

type paymentTerms struct {
	DueDate dateTimeWrap `xml:"ram:DueDateDateTime"`
}

// paymentMeans is a CII payment-instructions entry (BG-16).
type paymentMeans struct {
	TypeCode     string                 `xml:"ram:TypeCode"` // BT-81
	PayeeAccount *payeeFinancialAccount `xml:"ram:PayeePartyCreditorFinancialAccount,omitempty"`
	PayeeInst    *financialInstitution  `xml:"ram:PayeeSpecifiedCreditorFinancialInstitution,omitempty"`
}

type payeeFinancialAccount struct {
	IBANID string `xml:"ram:IBANID"` // BT-84
}

type financialInstitution struct {
	BICID string `xml:"ram:BICID"` // BT-86
}

type monetarySummation struct {
	LineTotalAmount     string       `xml:"ram:LineTotalAmount"`
	TaxBasisTotalAmount string       `xml:"ram:TaxBasisTotalAmount"`
	TaxTotalAmount      currencyAmount `xml:"ram:TaxTotalAmount"`
	GrandTotalAmount    string       `xml:"ram:GrandTotalAmount"`
	DuePayableAmount    string       `xml:"ram:DuePayableAmount"`
}

type currencyAmount struct {
	CurrencyID string `xml:"currencyID,attr"`
	Value      string `xml:",chardata"`
}
