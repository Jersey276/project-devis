package tests

import (
	"testing"

	"project-devis-invoice/actions"
	"project-devis-invoice/actions/codes"
)

func TestScheduleEligible(t *testing.T) {
	cases := map[string]bool{
		"VALID":     true,
		"DRAFT":     false,
		"NEGOCIATE": false,
		"DENIED":    false,
		"":          false,
	}
	for status, want := range cases {
		if got := actions.ScheduleEligibleForTest(status); got != want {
			t.Errorf("ScheduleEligible(%q) = %v; want %v", status, got, want)
		}
	}
}

func TestValidateMonthSelection_Valid(t *testing.T) {
	// Months 1 and 2 of a 6-month schedule, none billed yet.
	if got := actions.ValidateMonthSelectionForTest([]int32{1, 2}, 6, nil); got != codes.Success {
		t.Fatalf("got code %d; want Success (%d)", got, codes.Success)
	}
}

func TestValidateMonthSelection_Empty(t *testing.T) {
	if got := actions.ValidateMonthSelectionForTest(nil, 6, nil); got != codes.InvalidInput {
		t.Fatalf("empty selection got %d; want InvalidInput (%d)", got, codes.InvalidInput)
	}
}

func TestValidateMonthSelection_OutOfRange(t *testing.T) {
	// Month 7 does not exist in a 6-month schedule.
	if got := actions.ValidateMonthSelectionForTest([]int32{7}, 6, nil); got != codes.InvalidInput {
		t.Fatalf("out-of-range got %d; want InvalidInput (%d)", got, codes.InvalidInput)
	}
	// Month 0 is below the 1-based range.
	if got := actions.ValidateMonthSelectionForTest([]int32{0}, 6, nil); got != codes.InvalidInput {
		t.Fatalf("month 0 got %d; want InvalidInput (%d)", got, codes.InvalidInput)
	}
}

func TestValidateMonthSelection_Duplicates(t *testing.T) {
	if got := actions.ValidateMonthSelectionForTest([]int32{2, 2}, 6, nil); got != codes.InvalidInput {
		t.Fatalf("duplicate months got %d; want InvalidInput (%d)", got, codes.InvalidInput)
	}
}

func TestValidateMonthSelection_AlreadyBilled(t *testing.T) {
	// Month 2 was already invoiced; re-billing it must be refused.
	if got := actions.ValidateMonthSelectionForTest([]int32{2, 3}, 6, []int32{1, 2}); got != codes.MonthsAlreadyBilled {
		t.Fatalf("already-billed got %d; want MonthsAlreadyBilled (%d)", got, codes.MonthsAlreadyBilled)
	}
}

func TestValidateMonthSelection_DisjointFromBilled(t *testing.T) {
	// Months 3,4 requested; 1,2 already billed — disjoint, so valid.
	if got := actions.ValidateMonthSelectionForTest([]int32{3, 4}, 6, []int32{1, 2}); got != codes.Success {
		t.Fatalf("disjoint selection got %d; want Success (%d)", got, codes.Success)
	}
}
