package actions

import "project-devis-invoice/actions/codes"

const scheduleStatusValid = "VALID"

// validateMonthSelection checks a requested set of schedule month indexes
// against the schedule duration and the months already billed by ISSUED
// invoices. It is the pure core of the from-schedule eligibility rules.
//
// Returns codes.Success when the selection is valid, otherwise:
//   - codes.InvalidInput        — empty selection, duplicates, or out-of-range month
//   - codes.MonthsAlreadyBilled — at least one requested month is already invoiced
//
func validateMonthSelection(requested []int32, durationMonths int32, alreadyBilled []int32) int32 {
	if len(requested) == 0 {
		return codes.InvalidInput
	}

	billed := make(map[int32]struct{}, len(alreadyBilled))
	for _, m := range alreadyBilled {
		billed[m] = struct{}{}
	}

	seen := make(map[int32]struct{}, len(requested))
	for _, m := range requested {
		if m < 1 || m > durationMonths {
			return codes.InvalidInput
		}
		if _, dup := seen[m]; dup {
			return codes.InvalidInput
		}
		seen[m] = struct{}{}
		if _, already := billed[m]; already {
			return codes.MonthsAlreadyBilled
		}
	}
	return codes.Success
}

// scheduleEligible reports whether a schedule in the given status can be billed.
// Only VALID schedules may produce invoices.
func scheduleEligible(status string) bool {
	return status == scheduleStatusValid
}
