package actions

import usersGrpc "project-devis-invoice/services/usersgrpc"

// PickDestinationTaxForTest exposes the unexported destination-rate selection
// used by OSS resolution to the external `tests` package. It returns the chosen
// tax's rate string and label, plus ok=false when no tax is available.
func PickDestinationTaxForTest(taxes []*usersGrpc.Tax) (rate, label string, ok bool) {
	t := pickDestinationTax(taxes)
	if t == nil {
		return "", "", false
	}
	return t.GetRate(), t.GetName(), true
}
