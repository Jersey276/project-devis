package actions

// schedulesBalanced reports whether the planned amounts per quote line match
// the expected amounts exactly — no line over- or under-planned.
func schedulesBalanced(expected, planned map[string]int64) bool {
	for lineID, exp := range expected {
		if planned[lineID] != exp {
			return false
		}
	}
	for lineID, pln := range planned {
		if _, ok := expected[lineID]; !ok && pln != 0 {
			return false
		}
	}
	return true
}
