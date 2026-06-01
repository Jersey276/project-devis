package codes

const (
	Success int32 = 0
	// NotFound: the quote itself does not exist (or is not owned by the caller).
	NotFound int32 = 3001
	// DependencyMissing: the quote exists but one of its referenced entities
	// (client, address) does not. Distinct from NotFound so the gateway can
	// render a different message.
	DependencyMissing int32 = 3005
	// QuoteRefused: the quote is in DROP (refused) state and cannot be exported.
	QuoteRefused  int32 = 3006
	InternalError int32 = 3003
	InvalidInput  int32 = 3004
)
