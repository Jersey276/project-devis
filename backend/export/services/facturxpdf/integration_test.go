package facturxpdf

import (
	"bytes"
	"context"
	"os"
	"testing"

	"project-devis-export/services/gotenberg"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// TestAssemble_EndToEnd_RealGotenberg drives the full chain against a live
// Gotenberg: HTML -> PDF/A-3b -> Assemble -> Factur-X. Skipped unless
// FACTURX_INTEGRATION is set; GOTENBERG_ADDRESS must point at the service
// (e.g. run from a container on the backend_default network).
func TestAssemble_EndToEnd_RealGotenberg(t *testing.T) {
	if os.Getenv("FACTURX_INTEGRATION") == "" {
		t.Skip("set FACTURX_INTEGRATION to run the end-to-end Gotenberg test")
	}
	addr := os.Getenv("GOTENBERG_ADDRESS")
	if addr == "" {
		addr = "http://gotenberg:3000"
	}

	gt := gotenberg.New(addr)
	html := []byte(`<!doctype html><html><head><meta charset="utf-8"><title>Facture</title></head>` +
		`<body><h1>FACTURE 2026-0001</h1><p>Total TTC : 120,00 €</p></body></html>`)

	pdfA3, err := gt.ConvertPDFA3(context.Background(), html)
	if err != nil {
		t.Fatalf("ConvertPDFA3: %v", err)
	}
	if !bytes.Contains(pdfA3, []byte("<pdfaid:part>3")) {
		t.Fatal("Gotenberg output is not PDF/A-3")
	}

	out, err := Assemble(pdfA3, []byte(sampleXML))
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	// Attachment present.
	atts, err := api.Attachments(bytes.NewReader(out), model.NewDefaultConfiguration())
	if err != nil {
		t.Fatalf("attachments: %v", err)
	}
	if len(atts) != 1 || atts[0].FileName != attachmentName {
		t.Fatalf("attachments = %+v; want one %q", atts, attachmentName)
	}

	// Still PDF/A AND now Factur-X.
	xmp := extractMetadata(t, out)
	if !bytes.Contains(xmp, []byte("<pdfaid:part>3")) {
		t.Error("assembled PDF lost its PDF/A-3 identification")
	}
	if !bytes.Contains(xmp, []byte("urn:factur-x")) {
		t.Error("assembled PDF has no Factur-X XMP block")
	}
}
