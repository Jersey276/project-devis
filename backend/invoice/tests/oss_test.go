package tests

import (
	"testing"

	"project-devis-invoice/actions"
	usersGrpc "project-devis-invoice/services/usersgrpc"
)

func tax(rate, name string, isDefault bool) *usersGrpc.Tax {
	return &usersGrpc.Tax{Rate: rate, Name: name, IsDefault: isDefault}
}

func TestPickDestinationTax(t *testing.T) {
	cases := []struct {
		name     string
		taxes    []*usersGrpc.Tax
		wantRate string
		wantOK   bool
	}{
		{
			name:   "empty -> none",
			taxes:  nil,
			wantOK: false,
		},
		{
			name:     "single rate",
			taxes:    []*usersGrpc.Tax{tax("19", "USt 19%", false)},
			wantRate: "19",
			wantOK:   true,
		},
		{
			name: "default flag wins over higher rate",
			taxes: []*usersGrpc.Tax{
				tax("25", "Top rate", false),
				tax("19", "Standard", true),
			},
			wantRate: "19",
			wantOK:   true,
		},
		{
			name: "no default -> highest rate",
			taxes: []*usersGrpc.Tax{
				tax("7", "Reduced", false),
				tax("19", "Standard", false),
			},
			wantRate: "19",
			wantOK:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rate, _, ok := actions.PickDestinationTaxForTest(tc.taxes)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v; want %v", ok, tc.wantOK)
			}
			if ok && rate != tc.wantRate {
				t.Errorf("rate = %q; want %q", rate, tc.wantRate)
			}
		})
	}
}
