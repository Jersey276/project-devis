package tests

import (
	"testing"
	"time"

	"project-devis-invoice/actions"
)

func goldenDoc() actions.SealableDocForTest {
	issued, _ := time.Parse(time.RFC3339, "2026-06-14T10:30:00Z")
	return actions.SealableDocForTest{
		UserID:   "user-golden",
		DocType:  "INVOICE",
		Number:   "2026-0001",
		IssuedAt: issued,
		TotalHT:  10000,
		TotalVAT: 2000,
		TotalTTC: 12000,
		Lines: []actions.SealLineForTest{
			{Name: "Prestation A", Quantity: "1", Unit: "u", UnitPriceCents: 10000, LineHTCents: 10000, TaxRate: "20", TaxLabel: "TVA 20%"},
		},
	}
}

// TestComputeContentHash_Golden pins the canonical serialization. If this breaks,
// the wire format changed and every existing seal is invalidated — do not "fix"
// the expected value without a re-seal migration.
func TestComputeContentHash_Golden(t *testing.T) {
	const want = "d7c7a2ebe647b52a005bf31ab009bd3e2db12dc873fef154c289f694bae20ff6"
	if got := actions.ComputeContentHashForTest(goldenDoc()); got != want {
		t.Fatalf("content hash = %s; want golden %s (frozen wire format changed?)", got, want)
	}
}

func TestComputeChainHash_GenesisGolden(t *testing.T) {
	content := actions.ComputeContentHashForTest(goldenDoc())
	const want = "55d6477b847e4ff5862882aa3e59317251710733102760c06af8bf4c3c48607c"
	if got := actions.ComputeChainHashForTest(actions.GenesisHashForTest(), content, 0); got != want {
		t.Fatalf("genesis chain hash = %s; want golden %s", got, want)
	}
}

func TestComputeContentHash_Stable(t *testing.T) {
	d := goldenDoc()
	if actions.ComputeContentHashForTest(d) != actions.ComputeContentHashForTest(d) {
		t.Fatal("content hash is not stable across calls")
	}
}

func TestComputeContentHash_FieldSensitivity(t *testing.T) {
	base := actions.ComputeContentHashForTest(goldenDoc())

	mutators := map[string]func(d *actions.SealableDocForTest){
		"totalTTC":  func(d *actions.SealableDocForTest) { d.TotalTTC++ },
		"vatExempt": func(d *actions.SealableDocForTest) { d.VatExempt = !d.VatExempt },
		"number":    func(d *actions.SealableDocForTest) { d.Number = "2026-0002" },
		"issuedAt":  func(d *actions.SealableDocForTest) { d.IssuedAt = d.IssuedAt.Add(time.Second) },
		"origin":    func(d *actions.SealableDocForTest) { d.OriginInvoiceNumber = "2026-0009" },
		"lineHT":    func(d *actions.SealableDocForTest) { d.Lines[0].LineHTCents++ },
		"lineName":  func(d *actions.SealableDocForTest) { d.Lines[0].Name = "Autre" },
		"taxRate":   func(d *actions.SealableDocForTest) { d.Lines[0].TaxRate = "10" },
	}
	for name, mut := range mutators {
		d := goldenDoc()
		mut(&d)
		if actions.ComputeContentHashForTest(d) == base {
			t.Errorf("mutating %s did not change the content hash", name)
		}
	}
}

func TestComputeChainHash_Sensitivity(t *testing.T) {
	content := actions.ComputeContentHashForTest(goldenDoc())
	base := actions.ComputeChainHashForTest(actions.GenesisHashForTest(), content, 0)

	if actions.ComputeChainHashForTest("ff", content, 0) == base {
		t.Error("changing prevHash did not change chain hash")
	}
	if actions.ComputeChainHashForTest(actions.GenesisHashForTest(), content, 1) == base {
		t.Error("changing index did not change chain hash")
	}
	if actions.ComputeChainHashForTest(actions.GenesisHashForTest(), "ff", 0) == base {
		t.Error("changing content did not change chain hash")
	}
}

func TestComputeContentHash_TimezoneIndependent(t *testing.T) {
	// The same instant in different zones must hash identically (UTC normalisation).
	paris, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Skip("tzdata unavailable")
	}
	utc := goldenDoc()
	parisDoc := goldenDoc()
	parisDoc.IssuedAt = utc.IssuedAt.In(paris)
	if actions.ComputeContentHashForTest(utc) != actions.ComputeContentHashForTest(parisDoc) {
		t.Fatal("same instant in different timezones produced different hashes")
	}
}
