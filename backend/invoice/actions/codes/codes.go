package codes

const (
	Success             int32 = 0
	NotFound            int32 = 1001
	AlreadyExists       int32 = 1002
	InvalidInput        int32 = 1003
	SourceNotEligible   int32 = 4001 // schedule not VALID / quote not validated
	QuoteHasSchedule    int32 = 4002 // a schedule exists; bill from the schedule
	InvoiceFinalized    int32 = 4003 // already ISSUED/CANCELLED — immutable
	MonthsAlreadyBilled int32 = 4004 // some requested months are already invoiced
	DependencyMissing   int32 = 4005 // a referenced client/address is missing

	CreditNoteLineAlreadyCredited int32 = 4006 // a selected line is already credited
	InvoiceNotIssued              int32 = 4007 // the invoice is not ISSUED/PAID
	CreditNoteNoLinesLeft         int32 = 4008 // total credit requested but nothing left
	SealError                     int32 = 4009 // failed to seal the document at emission
	OSSDestinationTaxMissing      int32 = 4010 // OSS applies but no tax configured for the client's country

	InternalError int32 = 2001
)
