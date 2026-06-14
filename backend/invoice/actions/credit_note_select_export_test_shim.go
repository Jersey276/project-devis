package actions

// ResolveCreditedPositionsForTest exposes the pure selection resolver to the
// tests package.
func ResolveCreditedPositionsForTest(requested, allPositions []int32, alreadyCredited []int32) (selected []int32, isTotal bool, code int32) {
	set := make(map[int32]struct{}, len(alreadyCredited))
	for _, p := range alreadyCredited {
		set[p] = struct{}{}
	}
	return resolveCreditedPositions(requested, allPositions, set)
}
