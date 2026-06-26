// Package pdp is the neutral seam to a French e-invoicing Plateforme Agréée (PA,
// ex-PDP). It isolates the platform contract from business logic: no proto, no
// DB. No real provider is wired (Iopole/Seqino shortlisted, no contract) — the
// default adapter is a no-op, with a programmable mock for tests.
package pdp

import "context"

// PlatformStatus is the PA-side status, kept distinct from the B3 lifecycle
// vocabulary so the mapping (mapping.go) stays explicit and provider-agnostic.
type PlatformStatus string

const (
	PlatformSubmitted PlatformStatus = "SUBMITTED" // accepted for deposit by the platform
	PlatformReceived  PlatformStatus = "RECEIVED"
	PlatformApproved  PlatformStatus = "APPROVED"
	PlatformRejected  PlatformStatus = "REJECTED"
	PlatformCollected PlatformStatus = "COLLECTED"
	PlatformUnknown   PlatformStatus = "UNKNOWN"
)

// SubmitInput carries the minimum a platform needs to deposit one issued
// invoice. Kept small on purpose; real adapters can read more later.
type SubmitInput struct {
	InvoiceID     string
	UserID        string
	InvoiceNumber string
}

// SubmitResult is the outcome of depositing one issued invoice.
type SubmitResult struct {
	SubmissionID string         // platform-assigned handle; "" for the no-op adapter
	Status       PlatformStatus // typically PlatformSubmitted on success
}

// Client is the neutral interface to a Plateforme Agréée.
type Client interface {
	// Submit deposits an already-issued invoice and returns the platform handle.
	Submit(ctx context.Context, in SubmitInput) (SubmitResult, error)
	// FetchStatus polls the current platform status for a prior submission.
	FetchStatus(ctx context.Context, submissionID string) (PlatformStatus, error)
}
