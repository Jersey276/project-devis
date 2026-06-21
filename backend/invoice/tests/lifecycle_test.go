package tests

import (
	"testing"

	"project-devis-invoice/actions"
)

func TestNextLifecycleAllowed(t *testing.T) {
	cases := []struct {
		current, target string
		want            bool
	}{
		// Happy path, one hop at a time.
		{"NONE", "DEPOSITED", true},
		{"DEPOSITED", "RECEIVED", true},
		{"RECEIVED", "APPROVED", true},
		{"APPROVED", "COLLECTED", true},

		// REJECTED reachable from any active non-terminal state.
		{"DEPOSITED", "REJECTED", true},
		{"RECEIVED", "REJECTED", true},
		{"APPROVED", "REJECTED", true},
		{"NONE", "REJECTED", false},

		// No skipping.
		{"NONE", "RECEIVED", false},
		{"NONE", "APPROVED", false},
		{"DEPOSITED", "APPROVED", false},
		{"DEPOSITED", "COLLECTED", false},
		{"RECEIVED", "COLLECTED", false},

		// No going backward.
		{"RECEIVED", "DEPOSITED", false},
		{"APPROVED", "RECEIVED", false},
		{"COLLECTED", "APPROVED", false},

		// Terminal states cannot move.
		{"REJECTED", "DEPOSITED", false},
		{"REJECTED", "COLLECTED", false},
		{"COLLECTED", "REJECTED", false},

		// No self-loops (keeps the append-only log duplicate-free).
		{"DEPOSITED", "DEPOSITED", false},
		{"COLLECTED", "COLLECTED", false},

		// Unknown / sentinel targets.
		{"DEPOSITED", "NONE", false},
		{"DEPOSITED", "UNKNOWN", false},
		{"UNKNOWN", "DEPOSITED", false},
	}

	for _, c := range cases {
		if got := actions.NextLifecycleAllowedForTest(c.current, c.target); got != c.want {
			t.Errorf("NextLifecycleAllowed(%q, %q) = %v; want %v", c.current, c.target, got, c.want)
		}
	}
}
