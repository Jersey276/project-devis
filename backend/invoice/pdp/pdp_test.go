package pdp

import (
	"context"
	"testing"
)

func TestToLifecycleStatus(t *testing.T) {
	cases := []struct {
		in     PlatformStatus
		want   string
		wantOK bool
	}{
		{PlatformSubmitted, "DEPOSITED", true},
		{PlatformReceived, "RECEIVED", true},
		{PlatformApproved, "APPROVED", true},
		{PlatformRejected, "REJECTED", true},
		{PlatformCollected, "COLLECTED", true},
		{PlatformUnknown, "", false},
		{PlatformStatus("GARBAGE"), "", false},
	}
	for _, tc := range cases {
		got, ok := ToLifecycleStatus(tc.in)
		if got != tc.want || ok != tc.wantOK {
			t.Errorf("ToLifecycleStatus(%q) = (%q,%v); want (%q,%v)", tc.in, got, ok, tc.want, tc.wantOK)
		}
	}
}

func TestNoopDirectory_ResolvesEveryone(t *testing.T) {
	routing, err := NoopDirectory{}.Resolve(context.Background(), "12345678900011")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if routing.RoutingID != "" {
		t.Errorf("RoutingID=%q; want empty (no-op assigns none)", routing.RoutingID)
	}
}

func TestNoopClient_SubmitAcceptsLocally(t *testing.T) {
	res, err := NoopClient{}.Submit(context.Background(), SubmitInput{InvoiceID: "x"})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if res.Status != PlatformSubmitted {
		t.Errorf("status=%q; want SUBMITTED", res.Status)
	}
	if res.SubmissionID != "" {
		t.Errorf("SubmissionID=%q; want empty (no-op assigns none)", res.SubmissionID)
	}
}
