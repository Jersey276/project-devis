package iopole

import "project-devis-invoice/pdp"

// statusMap collapses Iopole's 15 lifecycle codes onto the neutral PlatformStatus
// vocabulary (pdp/mapping.go then maps that onto the B3 lifecycle). Codes with no
// matching B3 step (DISPUTED, SUSPENDED) and any unknown code resolve to UNKNOWN,
// so the poller never invents a transition.
var statusMap = map[string]pdp.PlatformStatus{
	"SUBMITTED":          pdp.PlatformSubmitted,
	"ISSUED":             pdp.PlatformSubmitted,
	"RECEIVED":           pdp.PlatformReceived,
	"MADE_AVAILABLE":     pdp.PlatformReceived,
	"IN_HAND":            pdp.PlatformReceived,
	"APPROVED":           pdp.PlatformApproved,
	"PARTIALLY_APPROVED": pdp.PlatformApproved,
	"COMPLETED":          pdp.PlatformApproved,
	"PAYMENT_SENT":       pdp.PlatformCollected,
	"PAYMENT_RECEIVED":   pdp.PlatformCollected,
	"REJECTED":           pdp.PlatformRejected,
	"REFUSED":            pdp.PlatformRejected,
	"UNACCEPTABLE":       pdp.PlatformRejected,
}

// mapStatus translates an Iopole status code to a PlatformStatus, defaulting to
// UNKNOWN for unmapped or unrecognised codes.
func mapStatus(code string) pdp.PlatformStatus {
	if s, ok := statusMap[code]; ok {
		return s
	}
	return pdp.PlatformUnknown
}
