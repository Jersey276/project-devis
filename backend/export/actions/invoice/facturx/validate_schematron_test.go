package facturx

import (
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	invoicepb "project-devis-export/services/invoice"
)

// Schematron is the EN 16931 *business-rule* gate, layered on top of the XSD
// structural gate (validate_xsd_test.go). XSD proves the CII is well-formed;
// Schematron proves it satisfies the BR-* / BR-CO-* / BR-S-* rules a PDP applies.
//
// The compiled rules ship as an XSLT 2.0 stylesheet (queryBinding="xslt2"), so —
// unlike the XSD check — xmllint/xsltproc cannot run them; we run them with
// Saxon-HE (vendored jar) and parse the resulting SVRL report. See
// schematron/README.md for provenance.
const (
	// ciiSchematronXSLT is the compiled EN 16931 CII Schematron, relative to this
	// package directory.
	ciiSchematronXSLT = "schematron/en16931/EN16931-CII-validation.xslt"
	// saxonJar is the vendored Saxon-HE processor (self-contained 10.x).
	saxonJar = "schematron/saxon/saxon-he-10.9.jar"
)

// TestSchematron_EN16931 proves every document we build satisfies the EN 16931
// business rules, not just the XSD structure. It shells out to Saxon (a JRE);
// when no usable JRE is reachable it skips, so a dev box without Java stays green
// while CI (which installs Temurin) does the real check.
func TestSchematron_EN16931(t *testing.T) {
	validate := saxonValidator(t)

	cases := []struct {
		name  string
		build func() ([]byte, error)
	}{
		{"invoice", func() ([]byte, error) { return Build(sampleInvoice()) }},
		{"invoice_vat_exempt", func() ([]byte, error) { return Build(exemptInvoice()) }},
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
			doc, err := tc.build()
			if err != nil {
				t.Fatalf("build: %v", err)
			}
			svrl, err := validate(t, doc)
			if err != nil {
				t.Fatalf("run schematron: %v", err)
			}
			if fatals := fatalAsserts(t, svrl); len(fatals) > 0 {
				t.Errorf("EN 16931 Schematron failed:\n%s", strings.Join(fatals, "\n"))
			}
		})
	}
}

// TestSchematron_RejectsInvalid guards the guard: a document that breaks a
// business rule (here BR-CO-15: grand total must equal tax basis + tax) must
// raise a fatal assert, proving the harness discriminates and is not silently
// passing everything.
func TestSchematron_RejectsInvalid(t *testing.T) {
	validate := saxonValidator(t)

	doc, err := Build(sampleInvoice())
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	// Corrupt the grand total so BR-CO-15 (175.00 = 150.00 + 25.00) no longer holds.
	broken := strings.Replace(string(doc),
		"<ram:GrandTotalAmount>175.00</ram:GrandTotalAmount>",
		"<ram:GrandTotalAmount>999.00</ram:GrandTotalAmount>", 1)
	if broken == string(doc) {
		t.Fatal("test setup: expected substitution did not happen")
	}
	svrl, err := validate(t, []byte(broken))
	if err != nil {
		t.Fatalf("run schematron: %v", err)
	}
	if fatals := fatalAsserts(t, svrl); len(fatals) == 0 {
		t.Error("rule-breaking CII XML unexpectedly passed Schematron")
	}
}

// saxonValidator returns a function that runs the CII Schematron XSLT over a
// document and returns the SVRL report. It resolves a runner once: a native JRE
// (11+) on PATH, or — on Windows — java inside the Debian WSL distro. It skips
// the whole test if neither is available.
func saxonValidator(t *testing.T) func(*testing.T, []byte) ([]byte, error) {
	t.Helper()
	xsltAbs, err := filepath.Abs(ciiSchematronXSLT)
	if err != nil {
		t.Fatalf("resolve xslt path: %v", err)
	}
	jarAbs, err := filepath.Abs(saxonJar)
	if err != nil {
		t.Fatalf("resolve saxon jar path: %v", err)
	}

	if nativeJavaUsable() {
		return func(t *testing.T, doc []byte) ([]byte, error) {
			return runSaxon(t, nil, jarAbs, xsltAbs, doc, false)
		}
	}
	if runtime.GOOS == "windows" && wslHasJava() {
		return func(t *testing.T, doc []byte) ([]byte, error) {
			return runSaxon(t, []string{"wsl.exe", "-d", "Debian", "--"}, jarAbs, xsltAbs, doc, true)
		}
	}
	t.Skip("no JRE 11+ available (install a JRE to run EN 16931 Schematron conformance)")
	return nil
}

// nativeJavaUsable reports whether a java on PATH is recent enough for Saxon-HE
// (Java 11+). An old JRE (e.g. Java 8) is treated as unusable so the resolver
// falls through to WSL/skip instead of crashing on a class-version error.
func nativeJavaUsable() bool {
	if _, err := exec.LookPath("java"); err != nil {
		return false
	}
	return javaMajor(javaVersionOutput(nil)) >= 11
}

func wslHasJava() bool {
	if !wslHasCommand("java") {
		return false
	}
	return javaMajor(javaVersionOutput([]string{"wsl.exe", "-d", "Debian", "--"})) >= 11
}

func wslHasCommand(name string) bool {
	out, err := exec.Command("wsl.exe", "-d", "Debian", "--", "sh", "-c", "command -v "+name).Output()
	return err == nil && len(strings.TrimSpace(string(out))) > 0
}

// javaVersionOutput returns the combined output of `java -version` (which java
// prints to stderr), optionally through a launcher (wsl.exe ...).
func javaVersionOutput(launcher []string) string {
	args := append(append([]string{}, launcher...), "java", "-version")
	out, _ := exec.Command(args[0], args[1:]...).CombinedOutput()
	return string(out)
}

// javaMajor extracts the major version from a `java -version` banner, handling
// both legacy ("1.8.0_xxx") and modern ("21.0.1") version strings. Returns 0 if
// it cannot be parsed.
func javaMajor(banner string) int {
	i := strings.IndexByte(banner, '"')
	if i < 0 {
		return 0
	}
	rest := banner[i+1:]
	j := strings.IndexByte(rest, '"')
	if j < 0 {
		return 0
	}
	ver := rest[:j]
	parts := strings.Split(ver, ".")
	if len(parts) == 0 {
		return 0
	}
	if parts[0] == "1" && len(parts) > 1 { // legacy 1.8 -> 8
		return atoiSafe(parts[1])
	}
	return atoiSafe(parts[0])
}

func atoiSafe(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}

// runSaxon writes the document to a temp file, runs the Schematron XSLT over it
// with Saxon, and returns the SVRL output. When wsl is true the jar, stylesheet,
// input and output host paths are translated to WSL mount paths so java inside
// Debian can read them.
func runSaxon(t *testing.T, launcher []string, jarAbs, xsltAbs string, doc []byte, wsl bool) ([]byte, error) {
	t.Helper()
	dir := t.TempDir()
	docFile := filepath.Join(dir, "doc.xml")
	svrlFile := filepath.Join(dir, "report.svrl")
	if err := os.WriteFile(docFile, doc, 0o644); err != nil {
		t.Fatalf("write temp xml: %v", err)
	}

	jarArg, xsltArg, docArg, svrlArg := jarAbs, xsltAbs, docFile, svrlFile
	if wsl {
		jarArg = toWSLPath(t, jarAbs)
		xsltArg = toWSLPath(t, xsltAbs)
		docArg = toWSLPath(t, docFile)
		svrlArg = toWSLPath(t, svrlFile)
	}

	args := append([]string{}, launcher...)
	args = append(args, "java", "-jar", jarArg,
		"-xsl:"+xsltArg, "-s:"+docArg, "-o:"+svrlArg)
	cmd := exec.Command(args[0], args[1:]...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("saxon: %v\n%s", err, strings.TrimSpace(string(out)))
	}

	svrl, err := os.ReadFile(svrlFile)
	if err != nil {
		t.Fatalf("read svrl: %v", err)
	}
	return svrl, nil
}

// SVRL (Schematron Validation Report Language) structures. Elements live in the
// http://purl.oclc.org/dsdl/svrl namespace; we match on local name only, which
// ignores Saxon's prefix choice (the same pragmatism the XSD tests use for CII).
type svrlOutput struct {
	XMLName xml.Name     `xml:"schematron-output"`
	Failed  []svrlAssert `xml:"failed-assert"`
}

type svrlAssert struct {
	Test     string `xml:"test,attr"`
	Location string `xml:"location,attr"`
	Flag     string `xml:"flag,attr"`
	Role     string `xml:"role,attr"`
	Text     string `xml:"text"`
}

// fatalAsserts parses the SVRL and returns readable messages for every blocking
// failed assert. EN 16931 marks rules fatal|warning via flag (and sometimes
// role); warnings are logged but do not fail the build.
func fatalAsserts(t *testing.T, svrl []byte) []string {
	t.Helper()
	var out svrlOutput
	if err := xml.Unmarshal(svrl, &out); err != nil {
		t.Fatalf("parse svrl: %v\n%s", err, string(svrl))
	}
	var fatals []string
	for _, a := range out.Failed {
		msg := fmt.Sprintf("%s @ %s", strings.TrimSpace(a.Text), a.Location)
		if isWarning(a) {
			t.Logf("schematron warning: %s", msg)
			continue
		}
		fatals = append(fatals, msg)
	}
	return fatals
}

// isWarning reports whether a failed assert is non-blocking. The EN 16931 rules
// set flag="warning" (or role="warning") on advisory checks; everything else
// (flag="fatal", role="error", or unset) is treated as blocking.
func isWarning(a svrlAssert) bool {
	return strings.EqualFold(a.Flag, "warning") || strings.EqualFold(a.Role, "warning")
}
