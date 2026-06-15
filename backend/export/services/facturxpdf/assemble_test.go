package facturxpdf

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

const sampleXML = `<?xml version="1.0" encoding="UTF-8"?>` +
	`<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100">` +
	`<rsm:ExchangedDocument><ram:ID>2026-0001</ram:ID></rsm:ExchangedDocument>` +
	`</rsm:CrossIndustryInvoice>`

func loadSamplePDF(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "sample_pdfa3.pdf"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return b
}

func TestAssemble_EmbedsFacturX(t *testing.T) {
	out, err := Assemble(loadSamplePDF(t), []byte(sampleXML))
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("empty output")
	}

	// 1. The XML is embedded as factur-x.xml (api.Attachments validates and
	//    populates the EmbeddedFiles name tree, unlike a raw ReadContext).
	atts, err := api.Attachments(bytes.NewReader(out), model.NewDefaultConfiguration())
	if err != nil {
		t.Fatalf("list attachments: %v", err)
	}
	if len(atts) != 1 || atts[0].FileName != attachmentName {
		t.Fatalf("attachments = %+v; want one %q", atts, attachmentName)
	}

	// 2. Catalog /AF references a Filespec with /AFRelationship Alternative.
	ctx, err := api.ReadContext(bytes.NewReader(out), model.NewDefaultConfiguration())
	if err != nil {
		t.Fatalf("re-read assembled PDF: %v", err)
	}
	root, err := ctx.Catalog()
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	af, found := root.Find("AF")
	if !found {
		t.Fatal("catalog has no /AF array")
	}
	afArr, err := ctx.DereferenceArray(af)
	if err != nil {
		t.Fatalf("deref /AF: %v", err)
	}
	if len(afArr) != 1 {
		t.Fatalf("/AF length = %d; want 1", len(afArr))
	}
	spec, err := ctx.DereferenceDict(afArr[0])
	if err != nil {
		t.Fatalf("deref filespec: %v", err)
	}
	if rel := spec["AFRelationship"]; rel != types.Name("Alternative") {
		t.Errorf("/AFRelationship = %v; want Alternative", rel)
	}
}

func TestAssemble_XMPHasFacturxAndKeepsPDFA(t *testing.T) {
	out, err := Assemble(loadSamplePDF(t), []byte(sampleXML))
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	xmp := extractMetadata(t, out)

	for _, want := range []string{
		"urn:factur-x:pdfa:CrossIndustryDocument:invoice:1p0#",
		"<fx:DocumentType>INVOICE</fx:DocumentType>",
		"<fx:DocumentFileName>factur-x.xml</fx:DocumentFileName>",
		facturxConformanceLevel,
		"<pdfaid:part>3", // PDF/A identification preserved
	} {
		if !bytes.Contains(xmp, []byte(want)) {
			t.Errorf("XMP missing %q", want)
		}
	}
}

func TestAssemble_RejectsEmptyInputs(t *testing.T) {
	if _, err := Assemble(nil, []byte(sampleXML)); err == nil {
		t.Error("expected error for empty PDF")
	}
	if _, err := Assemble(loadSamplePDF(t), nil); err == nil {
		t.Error("expected error for empty XML")
	}
}

func TestAssemble_RejectsInvalidPDF(t *testing.T) {
	if _, err := Assemble([]byte("not a pdf"), []byte(sampleXML)); err == nil {
		t.Error("expected error for invalid PDF bytes")
	}
}

// extractMetadata reads the decoded /Metadata XMP stream from a PDF.
func extractMetadata(t *testing.T, pdf []byte) []byte {
	t.Helper()
	ctx, err := api.ReadContext(bytes.NewReader(pdf), model.NewDefaultConfiguration())
	if err != nil {
		t.Fatalf("read PDF: %v", err)
	}
	root, err := ctx.Catalog()
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	md, found := root.Find("Metadata")
	if !found {
		t.Fatal("no /Metadata")
	}
	sd, _, err := ctx.DereferenceStreamDict(md)
	if err != nil {
		t.Fatalf("deref metadata: %v", err)
	}
	if err := sd.Decode(); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	return sd.Content
}
