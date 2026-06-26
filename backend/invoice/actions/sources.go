package actions

import "project-devis-invoice/actions/codes"

const scheduleStatusValid = "VALID"

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

func scheduleEligible(status string) bool {
	return status == scheduleStatusValid
}
