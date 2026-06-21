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
