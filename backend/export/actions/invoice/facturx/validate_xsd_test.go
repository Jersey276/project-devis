package facturx

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	invoicepb "project-devis-export/services/invoice"
)

// ciiRootSchema is the EN 16931 / Factur-X CII root XSD, relative to this
// package directory. The whole schemas/uncefact tree must travel with it.
const ciiRootSchema = "schemas/uncefact/data/standard/CrossIndustryInvoice_100pD16B.xsd"

// TestXSD_Conformance proves the generator emits structurally valid CII: every
// document we build must validate against the official UN/CEFACT D16B schema.
// It shells out to xmllint (libxml2); when no xmllint is reachable it skips, so
// a dev box without libxml2-utils stays green while CI (which installs it) does
// the real check.
func TestXSD_Conformance(t *testing.T) {
	validate := xmllintValidator(t)

	cases := []struct {
		name  string
		build func() ([]byte, error)
	}{
		{"invoice", func() ([]byte, error) { return Build(sampleInvoice()) }},
		{"invoice_vat_exempt", func() ([]byte, error) {
			in := sampleInvoice()
			in.VatExempt = true
			in.VatBreakdown = nil
			in.TotalVatCents = 0
			in.TotalTtcCents = in.TotalHtCents
			return Build(in)
		}},
		{"invoice_client_b2c", func() ([]byte, error) {
			in := sampleInvoice()
			in.Client = &invoicepb.InvoiceParty{FirstName: "Jean", LastName: "Dupont", ZipCode: "75002", City: "Paris"}
			return Build(in)
		}},
		{"invoice_oss", func() ([]byte, error) { return Build(ossInvoice()) }},
		{"invoice_payment_means", func() ([]byte, error) {
			in := sampleInvoice()
			in.Issuer.Iban = "FR7630006000011234567890189"
			in.Issuer.Bic = "BNPAFRPP"
			return Build(in)
		}},
		{"invoice_siret", func() ([]byte, error) {
			in := sampleInvoice()
			in.Issuer.Siret = "12345678200019"
			in.Client.Siret = "98765432100025"
			return Build(in)
		}},
		{"credit_note", func() ([]byte, error) { return BuildCreditNote(sampleCreditNote()) }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			xml, err := tc.build()
			if err != nil {
				t.Fatalf("build: %v", err)
			}
			if err := validate(t, xml); err != nil {
				t.Errorf("CII XSD validation failed:\n%v", err)
			}
		})
	}
}

// TestXSD_RejectsInvalid guards the guard: a deliberately corrupted document
// must fail XSD validation, proving the harness actually discriminates (and is
// not silently passing everything).
func TestXSD_RejectsInvalid(t *testing.T) {
	validate := xmllintValidator(t)

	xml, err := Build(sampleInvoice())
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	// Rename a mandatory element to something the schema does not expect.
	broken := strings.Replace(string(xml),
		"<ram:TypeCode>380</ram:TypeCode>",
		"<ram:NotARealElement>380</ram:NotARealElement>", 1)
	if broken == string(xml) {
		t.Fatal("test setup: expected substitution did not happen")
	}
	if err := validate(t, []byte(broken)); err == nil {
		t.Error("corrupted CII XML unexpectedly passed XSD validation")
	}
}

// xmllintValidator returns a function that validates XML against the CII root
// schema. It resolves xmllint once: a native binary on PATH, or — on Windows —
// xmllint inside the Debian WSL distro. It skips the whole test if neither is
// available.
func xmllintValidator(t *testing.T) func(*testing.T, []byte) error {
	t.Helper()
	schemaAbs, err := filepath.Abs(ciiRootSchema)
	if err != nil {
		t.Fatalf("resolve schema path: %v", err)
	}

	if _, err := exec.LookPath("xmllint"); err == nil {
		return func(t *testing.T, xml []byte) error {
			return runXmllint(t, nil, schemaAbs, xml, false)
		}
	}
	if runtime.GOOS == "windows" {
		if wslHasXmllint() {
			return func(t *testing.T, xml []byte) error {
				return runXmllint(t, []string{"wsl.exe", "-d", "Debian", "--"}, schemaAbs, xml, true)
			}
		}
	}
	t.Skip("xmllint not available (install libxml2-utils to run CII XSD conformance)")
	return nil
}

func wslHasXmllint() bool {
	out, err := exec.Command("wsl.exe", "-d", "Debian", "--", "sh", "-c", "command -v xmllint").Output()
	return err == nil && len(strings.TrimSpace(string(out))) > 0
}

// runXmllint writes the XML to a temp file next to the schema dir and validates
// it. When wsl is true, the schema and XML host paths are translated to WSL
// mount paths (/mnt/c/...) so xmllint inside Debian can read them.
func runXmllint(t *testing.T, launcher []string, schemaAbs string, xml []byte, wsl bool) error {
	t.Helper()
	xmlFile := filepath.Join(t.TempDir(), "doc.xml")
	if err := os.WriteFile(xmlFile, xml, 0o644); err != nil {
		t.Fatalf("write temp xml: %v", err)
	}

	schemaArg, xmlArg := schemaAbs, xmlFile
	if wsl {
		schemaArg = toWSLPath(t, schemaAbs)
		xmlArg = toWSLPath(t, xmlFile)
	}

	args := append([]string{}, launcher...)
	args = append(args, "xmllint", "--noout", "--schema", schemaArg, xmlArg)
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &xmllintError{output: strings.TrimSpace(string(out))}
	}
	return nil
}

func toWSLPath(t *testing.T, winPath string) string {
	t.Helper()
	// wslpath drops backslashes; feed it forward slashes (-a forces absolute).
	slashed := filepath.ToSlash(winPath)
	out, err := exec.Command("wsl.exe", "-d", "Debian", "--", "wslpath", "-a", slashed).Output()
	if err != nil {
		t.Fatalf("wslpath %q: %v", slashed, err)
	}
	return strings.TrimSpace(string(out))
}

type xmllintError struct{ output string }

func (e *xmllintError) Error() string { return e.output }
