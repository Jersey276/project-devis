package pdp

import "context"

// Directory is the neutral seam to the French e-invoicing central directory
// (annuaire DGFiP/AIFE): it resolves a recipient SIRET to the platform that
// recipient receives invoices on, plus the routing handle the platform expects.
// Like Client it carries no proto and no DB, so a real adapter (or the directory
// the contracted PA exposes) is swapped in later without touching business logic.
type Directory interface {
	// Resolve looks up the recipient identified by siret. ErrRecipientNotFound
	// means the directory has no reachable e-invoicing recipient for that SIRET,
	// which a caller must treat as "do not deposit".
	Resolve(ctx context.Context, siret string) (RecipientRouting, error)
}

// RecipientRouting is the frozen result of a directory lookup: where and how the
// recipient's platform wants the invoice delivered. RoutingID is the handle the
// PA uses to address the recipient; it may be empty for the no-op directory.
type RecipientRouting struct {
	RoutingID    string // recipient handle on its receiving platform; "" for no-op
	PlatformName string // human label of the recipient's platform, for the audit trail
}

// ErrRecipientNotFound is returned by Resolve when the directory has no
// e-invoicing recipient for the given SIRET. It is a business outcome, not a
// transport error, so the deposit flow maps it to its own code.
var ErrRecipientNotFound = directoryError("recipient not found in directory")

type directoryError string

func (e directoryError) Error() string { return string(e) }
