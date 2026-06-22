package pdp

import "context"

// NoopClient is the default adapter: it accepts every deposit locally without
// calling any external platform, so the lifecycle works end-to-end before a real
// PA is contracted. It assigns no submission id.
type NoopClient struct{}

func (NoopClient) Submit(context.Context, SubmitInput) (SubmitResult, error) {
	return SubmitResult{Status: PlatformSubmitted}, nil
}

func (NoopClient) FetchStatus(context.Context, string) (PlatformStatus, error) {
	return PlatformUnknown, nil
}

// NoopDirectory is the default directory adapter: every recipient resolves
// successfully with an empty routing handle, so deposits work end-to-end before
// a real annuaire is wired. It never reports a recipient as not found — that
// guard only bites once a real directory is in place.
type NoopDirectory struct{}

func (NoopDirectory) Resolve(context.Context, string) (RecipientRouting, error) {
	return RecipientRouting{}, nil
}

// NoopReporter is the default e-reporting adapter (B5/C5): it accepts every report
// locally without calling any external platform, assigning no report id. So the
// feature works end-to-end and stays inert in production until a real PA is wired.
type NoopReporter struct{}

func (NoopReporter) SubmitReport(context.Context, SubmitReportInput) (SubmitReportResult, error) {
	return SubmitReportResult{Status: PlatformSubmitted}, nil
}

func (NoopReporter) FetchReportStatus(context.Context, string) (PlatformStatus, error) {
	return PlatformUnknown, nil
}
