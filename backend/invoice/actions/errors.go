package actions

import "project-devis-invoice/actions/codes"

const (
	CodeSuccess             = codes.Success
	CodeNotFound            = codes.NotFound
	CodeAlreadyExists       = codes.AlreadyExists
	CodeInvalidInput        = codes.InvalidInput
	CodeSourceNotEligible   = codes.SourceNotEligible
	CodeQuoteHasSchedule    = codes.QuoteHasSchedule
	CodeInvoiceFinalized    = codes.InvoiceFinalized
	CodeMonthsAlreadyBilled = codes.MonthsAlreadyBilled
	CodeDependencyMissing   = codes.DependencyMissing

	CodeCreditNoteLineAlreadyCredited = codes.CreditNoteLineAlreadyCredited
	CodeInvoiceNotIssued              = codes.InvoiceNotIssued
	CodeCreditNoteNoLinesLeft         = codes.CreditNoteNoLinesLeft

	CodeInternalError = codes.InternalError
)
