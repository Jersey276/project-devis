package pdp

// ToLifecycleStatus maps a platform status to a B3 lifecycle status (see
// actions/lifecycle.go). Returns ("", false) when there is no matching move
// (e.g. UNKNOWN), so a caller never invents a transition. The full set is mapped
// for forward-compatibility; this iteration only consumes SUBMITTED→DEPOSITED.
func ToLifecycleStatus(p PlatformStatus) (string, bool) {
	switch p {
	case PlatformSubmitted:
		return "DEPOSITED", true
	case PlatformReceived:
		return "RECEIVED", true
	case PlatformApproved:
		return "APPROVED", true
	case PlatformRejected:
		return "REJECTED", true
	case PlatformCollected:
		return "COLLECTED", true
	default:
		return "", false
	}
}
