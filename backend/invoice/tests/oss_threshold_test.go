package tests

import (
	"testing"

	"project-devis-invoice/actions"
	usersGrpc "project-devis-invoice/services/usersgrpc"
)

func country(code string, isEu bool) *usersGrpc.Country {
	return &usersGrpc.Country{Code: code, IsEu: isEu}
}

func TestOSSApplies(t *testing.T) {
	const threshold = actions.OSSThresholdCentsForTest
	de := country("DE", true)

	cases := []struct {
		name       string
		ossEnabled bool
		cumulative int64
		clientType string
		clientCty  *usersGrpc.Country
		want       bool
	}{
		{"below threshold, not opted in -> false", false, threshold - 1, "individual", de, false},
		{"below threshold, opted in -> true (anticipation)", true, 0, "individual", de, true},
		{"exactly at threshold, not opted in -> true (>=)", false, threshold, "individual", de, true},
		{"above threshold, not opted in -> true", false, threshold + 50000, "individual", de, true},
		{"above threshold but B2B -> false", false, threshold + 1, "business", de, false},
		{"above threshold but FR -> false", false, threshold + 1, "individual", country("FR", true), false},
		{"above threshold but non-EU -> false", false, threshold + 1, "individual", country("CH", false), false},
		{"nil country -> false", false, threshold + 1, "individual", nil, false},
		{"opted in but B2B -> false", true, 0, "business", de, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := actions.OSSAppliesForTest(tc.ossEnabled, tc.cumulative, tc.clientType, tc.clientCty)
			if got != tc.want {
				t.Errorf("ossApplies(enabled=%v, cum=%d, type=%q, %v) = %v; want %v",
					tc.ossEnabled, tc.cumulative, tc.clientType, tc.clientCty, got, tc.want)
			}
		})
	}
}
