package pdp

import "context"

// MockClient is a programmable adapter for tests: canned result/error plus a
// recorded call log. It lives in the package (not a _test.go file) so the
// cross-package tests/ suite can import it, matching the repo convention. Not
// used in production wiring.
type MockClient struct {
	SubmitResult SubmitResult
	SubmitErr    error
	StatusResult PlatformStatus
	StatusErr    error
	Submitted    []SubmitInput
}

func (m *MockClient) Submit(_ context.Context, in SubmitInput) (SubmitResult, error) {
	m.Submitted = append(m.Submitted, in)
	if m.SubmitErr != nil {
		return SubmitResult{}, m.SubmitErr
	}
	return m.SubmitResult, nil
}

func (m *MockClient) FetchStatus(context.Context, string) (PlatformStatus, error) {
	return m.StatusResult, m.StatusErr
}

// MockDirectory is a programmable directory adapter for tests: canned
// result/error plus a log of the SIRETs it was asked to resolve.
type MockDirectory struct {
	ResolveResult RecipientRouting
	ResolveErr    error
	Resolved      []string
}

func (m *MockDirectory) Resolve(_ context.Context, siret string) (RecipientRouting, error) {
	m.Resolved = append(m.Resolved, siret)
	if m.ResolveErr != nil {
		return RecipientRouting{}, m.ResolveErr
	}
	return m.ResolveResult, nil
}
