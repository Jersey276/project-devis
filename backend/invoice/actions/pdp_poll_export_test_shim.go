package actions

// ReconcileStepsForTest exposes reconcileSteps to the tests package.
func ReconcileStepsForTest(current, target string) []string {
	return reconcileSteps(current, target)
}
