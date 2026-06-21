package codes

const (
	Success             int32 = 0
	NotFound            int32 = 1001
	AlreadyExists       int32 = 1002
	InvalidInput        int32 = 1003
	SourceNotEligible   int32 = 4001
	QuoteHasSchedule    int32 = 4002
	InvoiceFinalized    int32 = 4003
	MonthsAlreadyBilled int32 = 4004
	DependencyMissing   int32 = 4005

	CreditNoteLineAlreadyCredited int32 = 4006
	InvoiceNotIssued              int32 = 4007
	CreditNoteNoLinesLeft         int32 = 4008
	SealError                     int32 = 4009
	OSSDestinationTaxMissing      int32 = 4010

	LifecycleTransitionInvalid int32 = 4011
	LifecycleRequiresIssued    int32 = 4012

	PDPSubmissionFailed     int32 = 4013
	RecipientNotInDirectory int32 = 4014

	InternalError int32 = 2001
)
