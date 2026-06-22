package actions

import (
	"strings"
	"testing"
	"time"

	"project-devis-invoice/pdp"
)

func TestReportScopeClause(t *testing.T) {
	tx, ok := reportScopeClause(pdp.ReportTransaction, "p")
	if !ok {
		t.Fatal("TRANSACTION scope: ok=false")
	}
	// Domestic B2C: individual + FR.
	if !strings.Contains(tx, "client_type = 'individual'") || !strings.Contains(tx, "client_country_code = 'FR'") {
		t.Errorf("TRANSACTION clause = %q; want individual + FR", tx)
	}

	cb, ok := reportScopeClause(pdp.ReportCrossBorderB2C, "p")
	if !ok {
		t.Fatal("CROSS_BORDER_B2C scope: ok=false")
	}
	// Intra-EU distance sales = the frozen OSS assiette: individual, not FR, counts flag.
	if !strings.Contains(cb, "client_country_code <> 'FR'") || !strings.Contains(cb, "counts_toward_oss_threshold") {
		t.Errorf("CROSS_BORDER_B2C clause = %q; want non-FR + assiette flag", cb)
	}

	if _, ok := reportScopeClause(pdp.ReportKind("GARBAGE"), "p"); ok {
		t.Error("unknown kind: ok=true; want false")
	}
}

func TestReportPeriodBounds(t *testing.T) {
	start, end := reportPeriodBounds(pdp.ReportPeriod{Year: 2099, Month: 2})
	wantStart := time.Date(2099, 2, 1, 0, 0, 0, 0, invoiceTZ)
	wantEnd := time.Date(2099, 3, 1, 0, 0, 0, 0, invoiceTZ)
	if !start.Equal(wantStart) || !end.Equal(wantEnd) {
		t.Errorf("bounds = [%s, %s); want [%s, %s)", start, end, wantStart, wantEnd)
	}
}

func TestReportKindFromString(t *testing.T) {
	for _, s := range []string{"TRANSACTION", "CROSS_BORDER_B2C"} {
		if _, ok := reportKindFromString(s); !ok {
			t.Errorf("reportKindFromString(%q): ok=false", s)
		}
	}
	if _, ok := reportKindFromString("nope"); ok {
		t.Error("reportKindFromString(nope): ok=true; want false")
	}
}
