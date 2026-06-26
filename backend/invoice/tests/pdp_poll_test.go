package tests

import (
	"reflect"
	"testing"

	"project-devis-invoice/actions"
)

func TestReconcileSteps(t *testing.T) {
	cases := []struct {
		current, target string
		want            []string
	}{
		// One cran forward.
		{"DEPOSITED", "RECEIVED", []string{"RECEIVED"}},
		{"RECEIVED", "APPROVED", []string{"APPROVED"}},
		{"APPROVED", "COLLECTED", []string{"COLLECTED"}},

		// Platform jumps ahead: B3 forbids skips, so the path is walked cran by cran.
		{"DEPOSITED", "APPROVED", []string{"RECEIVED", "APPROVED"}},
		{"DEPOSITED", "COLLECTED", []string{"RECEIVED", "APPROVED", "COLLECTED"}},
		{"RECEIVED", "COLLECTED", []string{"APPROVED", "COLLECTED"}},

		// REJECTED is a single direct move from any active state.
		{"DEPOSITED", "REJECTED", []string{"REJECTED"}},
		{"APPROVED", "REJECTED", []string{"REJECTED"}},

		// Already there: nothing to do.
		{"DEPOSITED", "DEPOSITED", nil},
		{"COLLECTED", "COLLECTED", nil},

		// Platform lags behind local state: no backward move.
		{"APPROVED", "RECEIVED", nil},
		{"COLLECTED", "DEPOSITED", nil},

		// Terminal current, or off-path inputs: no steps.
		{"COLLECTED", "REJECTED", nil},
		{"REJECTED", "COLLECTED", nil},
		{"DEPOSITED", "UNKNOWN", nil},
		{"NONE", "DEPOSITED", nil}, // NONE is reached by deposit, not by the poller.
	}

	for _, c := range cases {
		got := actions.ReconcileStepsForTest(c.current, c.target)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("reconcileSteps(%q, %q) = %v; want %v", c.current, c.target, got, c.want)
		}
	}
}
