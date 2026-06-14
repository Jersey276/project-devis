package tests

import (
	"testing"

	"project-devis-invoice/actions"
	"project-devis-invoice/actions/codes"
)

func TestResolveCreditedPositions_TotalOfRemainder(t *testing.T) {
	// Invoice has lines 0,1,2 ; line 1 already credited. Empty request = total
	// of the remainder → {0,2}, is_total true.
	sel, isTotal, code := actions.ResolveCreditedPositionsForTest(nil, []int32{0, 1, 2}, []int32{1})
	if code != codes.Success {
		t.Fatalf("code = %d; want Success", code)
	}
	if !isTotal {
		t.Fatalf("isTotal = false; want true")
	}
	if len(sel) != 2 || sel[0] != 0 || sel[1] != 2 {
		t.Fatalf("selected = %v; want [0 2]", sel)
	}
}

func TestResolveCreditedPositions_NothingLeft(t *testing.T) {
	// All lines already credited, empty request → CreditNoteNoLinesLeft.
	_, _, code := actions.ResolveCreditedPositionsForTest(nil, []int32{0, 1}, []int32{0, 1})
	if code != codes.CreditNoteNoLinesLeft {
		t.Fatalf("code = %d; want CreditNoteNoLinesLeft (%d)", code, codes.CreditNoteNoLinesLeft)
	}
}

func TestResolveCreditedPositions_PartialValid(t *testing.T) {
	sel, isTotal, code := actions.ResolveCreditedPositionsForTest([]int32{2, 0}, []int32{0, 1, 2}, nil)
	if code != codes.Success {
		t.Fatalf("code = %d; want Success", code)
	}
	if isTotal {
		t.Fatalf("isTotal = true; want false (partial)")
	}
	if len(sel) != 2 || sel[0] != 0 || sel[1] != 2 {
		t.Fatalf("selected = %v; want [0 2] (sorted)", sel)
	}
}

func TestResolveCreditedPositions_PartialIsTotalWhenAllSelected(t *testing.T) {
	_, isTotal, code := actions.ResolveCreditedPositionsForTest([]int32{0, 1}, []int32{0, 1}, nil)
	if code != codes.Success || !isTotal {
		t.Fatalf("code=%d isTotal=%v; want Success/true", code, isTotal)
	}
}

func TestResolveCreditedPositions_AlreadyCredited(t *testing.T) {
	_, _, code := actions.ResolveCreditedPositionsForTest([]int32{1}, []int32{0, 1, 2}, []int32{1})
	if code != codes.CreditNoteLineAlreadyCredited {
		t.Fatalf("code = %d; want CreditNoteLineAlreadyCredited (%d)", code, codes.CreditNoteLineAlreadyCredited)
	}
}

func TestResolveCreditedPositions_OutOfRange(t *testing.T) {
	_, _, code := actions.ResolveCreditedPositionsForTest([]int32{5}, []int32{0, 1, 2}, nil)
	if code != codes.InvalidInput {
		t.Fatalf("code = %d; want InvalidInput", code)
	}
}

func TestResolveCreditedPositions_Duplicate(t *testing.T) {
	_, _, code := actions.ResolveCreditedPositionsForTest([]int32{1, 1}, []int32{0, 1, 2}, nil)
	if code != codes.InvalidInput {
		t.Fatalf("code = %d; want InvalidInput", code)
	}
}
