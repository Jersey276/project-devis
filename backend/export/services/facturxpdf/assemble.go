// Package facturxpdf turns a PDF/A-3 visual invoice and its EN 16931 CII XML
// into a Factur-X document: the XML is embedded as the associated file
// "factur-x.xml" with /AFRelationship Alternative, referenced from the catalog
// /AF array, and the document XMP metadata is augmented with the urn:factur-x
// block so a Factur-X reader recognises the hybrid invoice.
package facturxpdf

import (
	"bytes"
	"fmt"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

const (
	// attachmentName is mandated by the Factur-X spec (exact file name).
	attachmentName = "factur-x.xml"
	// afRelationship: the PDF and the XML are equivalent representations
	// (Alternative) for the EN 16931 profile in the FR/DE context.
	afRelationship = "Alternative"
	attachmentDesc = "Factur-X EN 16931 invoice"

	// maxPDFBytes mirrors the gotenberg/gRPC cap so we never produce a PDF the
	// next hop would reject.
	maxPDFBytes = 8 * 1024 * 1024
)

// Assemble embeds xmlBytes into pdfBytes (a PDF/A-3b document) and returns the
// resulting Factur-X PDF. pdfBytes must already be PDF/A-3 (produced by
// gotenberg ConvertPDFA3); xmlBytes must be the CII invoice XML.
func Assemble(pdfBytes, xmlBytes []byte) ([]byte, error) {
	if len(pdfBytes) == 0 {
		return nil, fmt.Errorf("facturxpdf: empty PDF input")
	}
	if len(xmlBytes) == 0 {
		return nil, fmt.Errorf("facturxpdf: empty XML input")
	}

	conf := model.NewDefaultConfiguration()
	ctx, err := api.ReadContext(bytes.NewReader(pdfBytes), conf)
	if err != nil {
		return nil, fmt.Errorf("facturxpdf: read PDF: %w", err)
	}

	filespecRef, err := embedXML(ctx, xmlBytes)
	if err != nil {
		return nil, fmt.Errorf("facturxpdf: embed XML: %w", err)
	}
	if err := addAssociatedFile(ctx, filespecRef); err != nil {
		return nil, fmt.Errorf("facturxpdf: catalog /AF: %w", err)
	}
	if err := setFacturxXMP(ctx); err != nil {
		return nil, fmt.Errorf("facturxpdf: XMP: %w", err)
	}

	var out bytes.Buffer
	if err := api.WriteContext(ctx, &out); err != nil {
		return nil, fmt.Errorf("facturxpdf: write PDF: %w", err)
	}
	if out.Len() > maxPDFBytes {
		return nil, fmt.Errorf("facturxpdf: PDF exceeds %d bytes", maxPDFBytes)
	}
	return out.Bytes(), nil
}

// embedXML registers factur-x.xml in the EmbeddedFiles name tree and patches the
// Filespec dict with /AFRelationship. It mirrors model.AddAttachment but keeps
// the Filespec indirect reference so the catalog /AF array can point at it.
func embedXML(ctx *model.Context, xmlBytes []byte) (*types.IndirectRef, error) {
	xRefTable := ctx.XRefTable
	if err := xRefTable.LocateNameTree("EmbeddedFiles", true); err != nil {
		return nil, err
	}

	now := time.Now()
	a := model.Attachment{
		Reader:   bytes.NewReader(xmlBytes),
		ID:       attachmentName,
		FileName: attachmentName,
		Desc:     attachmentDesc,
		ModTime:  &now,
	}

	d, err := xRefTable.NewFileSpecDictForAttachment(a)
	if err != nil {
		return nil, err
	}
	// The associated-file relationship makes this a Factur-X attachment rather
	// than a plain embedded file.
	d["AFRelationship"] = types.Name(afRelationship)

	// PDF/A (ISO 19005-3, 6.8) requires the embedded file stream to declare its
	// MIME subtype; pdfcpu only sets /Type EmbeddedFile. text#2Fxml == text/xml.
	if err := setEmbeddedFileMIME(xRefTable, d); err != nil {
		return nil, err
	}

	ir, err := xRefTable.IndRefForNewObject(d)
	if err != nil {
		return nil, err
	}

	m := model.NameMap{a.ID: []types.Dict{d}}
	if err := xRefTable.Names["EmbeddedFiles"].Add(xRefTable, a.ID, *ir, m, []string{"F", "UF"}); err != nil {
		return nil, err
	}
	return ir, nil
}

// setEmbeddedFileMIME sets /Subtype (the MIME type) on the EmbeddedFile stream
// referenced by the Filespec's /EF /F entry. PDF/A requires it; pdfcpu omits it.
func setEmbeddedFileMIME(xRefTable *model.XRefTable, filespec types.Dict) error {
	ef, found := filespec.Find("EF")
	if !found {
		return fmt.Errorf("filespec has no /EF")
	}
	efDict, err := xRefTable.DereferenceDict(ef)
	if err != nil {
		return err
	}
	streamRef, found := efDict.Find("F")
	if !found {
		return fmt.Errorf("/EF has no /F")
	}
	sd, _, err := xRefTable.DereferenceStreamDict(streamRef)
	if err != nil {
		return err
	}
	// text#2Fxml is the PDF name-escaped form of "text/xml".
	sd.InsertName("Subtype", "text/xml")
	return nil
}

// addAssociatedFile appends the Filespec to the catalog /AF array (PDF/A-3
// associated files), creating it if absent.
func addAssociatedFile(ctx *model.Context, filespecRef *types.IndirectRef) error {
	root, err := ctx.Catalog()
	if err != nil {
		return err
	}
	if existing, found := root.Find("AF"); found {
		arr, err := ctx.DereferenceArray(existing)
		if err != nil {
			return err
		}
		root["AF"] = append(arr, *filespecRef)
		return nil
	}
	root["AF"] = types.Array{*filespecRef}
	return nil
}
