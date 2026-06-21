package actions

// NextLifecycleAllowedForTest exposes nextLifecycleAllowed to the tests package.
func NextLifecycleAllowedForTest(current, target string) bool {
	return nextLifecycleAllowed(current, target)
}
