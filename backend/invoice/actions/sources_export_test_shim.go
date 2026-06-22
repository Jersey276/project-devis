package actions

func ValidateMonthSelectionForTest(requested []int32, durationMonths int32, alreadyBilled []int32) int32 {
	return validateMonthSelection(requested, durationMonths, alreadyBilled)
}

func ScheduleEligibleForTest(status string) bool {
	return scheduleEligible(status)
}
