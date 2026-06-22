package sqlutil

import "testing"

func TestValidateSIRET(t *testing.T) {
	cases := []struct {
		name, siret, siren string
		wantErr            bool
	}{
		{"empty ok", "", "123456789", false},
		{"valid 14 digits", "12345678900012", "", false},
		{"valid with spaces", "123 456 789 00012", "", false},
		{"matches siren", "12345678900012", "123456789", false},
		{"too short", "123456789", "", true},
		{"non digit", "1234567890001A", "", true},
		{"siren mismatch", "98765432100012", "123456789", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ValidateSIRET(c.siret, c.siren)
			if (got != "") != c.wantErr {
				t.Fatalf("ValidateSIRET(%q,%q) = %q, wantErr=%v", c.siret, c.siren, got, c.wantErr)
			}
		})
	}
}
