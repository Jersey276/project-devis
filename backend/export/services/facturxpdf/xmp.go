package facturxpdf

import (
	"bytes"
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// facturxConformanceLevel is the Factur-X profile name carried in the XMP.
const facturxConformanceLevel = "EN 16931"

// facturxRDFBlock is the rdf:Description injected into the document XMP so a
// Factur-X reader can locate and trust the embedded XML. The fx namespace and
// keys are mandated by the Factur-X specification.
const facturxRDFBlock = `<rdf:Description rdf:about="" xmlns:fx="urn:factur-x:pdfa:CrossIndustryDocument:invoice:1p0#">` +
	`<fx:DocumentType>INVOICE</fx:DocumentType>` +
	`<fx:DocumentFileName>` + attachmentName + `</fx:DocumentFileName>` +
	`<fx:Version>1.0</fx:Version>` +
	`<fx:ConformanceLevel>` + facturxConformanceLevel + `</fx:ConformanceLevel>` +
	`</rdf:Description>`

// facturxExtensionSchema declares the fx namespace as a PDF/A extension schema.
// PDF/A (ISO 19005-3, 6.6.2.3) requires every custom XMP namespace to be
// described here, otherwise veraPDF rejects the fx:* properties. This block is
// fixed by the Factur-X specification.
const facturxExtensionSchema = `<rdf:Description rdf:about="" xmlns:pdfaExtension="http://www.aiim.org/pdfa/ns/extension/" xmlns:pdfaSchema="http://www.aiim.org/pdfa/ns/schema#" xmlns:pdfaProperty="http://www.aiim.org/pdfa/ns/property#">` +
	`<pdfaExtension:schemas><rdf:Bag><rdf:li rdf:parseType="Resource">` +
	`<pdfaSchema:schema>Factur-X PDFA Extension Schema</pdfaSchema:schema>` +
	`<pdfaSchema:namespaceURI>urn:factur-x:pdfa:CrossIndustryDocument:invoice:1p0#</pdfaSchema:namespaceURI>` +
	`<pdfaSchema:prefix>fx</pdfaSchema:prefix>` +
	`<pdfaSchema:property><rdf:Seq>` +
	`<rdf:li rdf:parseType="Resource"><pdfaProperty:name>DocumentFileName</pdfaProperty:name><pdfaProperty:valueType>Text</pdfaProperty:valueType><pdfaProperty:category>external</pdfaProperty:category><pdfaProperty:description>name of the embedded XML invoice file</pdfaProperty:description></rdf:li>` +
	`<rdf:li rdf:parseType="Resource"><pdfaProperty:name>DocumentType</pdfaProperty:name><pdfaProperty:valueType>Text</pdfaProperty:valueType><pdfaProperty:category>external</pdfaProperty:category><pdfaProperty:description>INVOICE</pdfaProperty:description></rdf:li>` +
	`<rdf:li rdf:parseType="Resource"><pdfaProperty:name>Version</pdfaProperty:name><pdfaProperty:valueType>Text</pdfaProperty:valueType><pdfaProperty:category>external</pdfaProperty:category><pdfaProperty:description>version of the Factur-X XML schema</pdfaProperty:description></rdf:li>` +
	`<rdf:li rdf:parseType="Resource"><pdfaProperty:name>ConformanceLevel</pdfaProperty:name><pdfaProperty:valueType>Text</pdfaProperty:valueType><pdfaProperty:category>external</pdfaProperty:category><pdfaProperty:description>conformance level of the Factur-X XML schema</pdfaProperty:description></rdf:li>` +
	`</rdf:Seq></pdfaSchema:property>` +
	`</rdf:li></rdf:Bag></pdfaExtension:schemas></rdf:Description>`

// setFacturxXMP merges the Factur-X rdf:Description into the document's existing
// XMP (the PDF/A metadata stream produced by Chromium), keeping the PDF/A
// identification intact. The metadata stream is rewritten unfiltered, as PDF/A
// requires the /Metadata stream not to be compressed.
func setFacturxXMP(ctx *model.Context) error {
	root, err := ctx.Catalog()
	if err != nil {
		return err
	}

	existing, found := root.Find("Metadata")
	if !found {
		return fmt.Errorf("document has no /Metadata (expected from PDF/A conversion)")
	}
	sd, _, err := ctx.DereferenceStreamDict(existing)
	if err != nil {
		return err
	}
	if err := sd.Decode(); err != nil {
		return fmt.Errorf("decode XMP stream: %w", err)
	}

	merged, err := injectFacturxBlock(sd.Content)
	if err != nil {
		return err
	}

	newSD, err := newMetadataStream(ctx, merged)
	if err != nil {
		return err
	}
	ir, err := ctx.IndRefForNewObject(*newSD)
	if err != nil {
		return err
	}
	root["Metadata"] = *ir
	return nil
}

// injectFacturxBlock inserts the Factur-X rdf:Description just before the
// closing </rdf:RDF> of the XMP packet. It is a targeted text insertion (no full
// XML parse) so the surrounding PDF/A XMP is preserved byte-for-byte.
func injectFacturxBlock(xmp []byte) ([]byte, error) {
	const marker = "</rdf:RDF>"
	idx := bytes.LastIndex(xmp, []byte(marker))
	if idx < 0 {
		return nil, fmt.Errorf("XMP packet has no %s element", marker)
	}
	var b bytes.Buffer
	b.Write(xmp[:idx])
	b.WriteString(facturxRDFBlock)
	b.WriteString(facturxExtensionSchema)
	b.Write(xmp[idx:])
	return b.Bytes(), nil
}

// newMetadataStream builds an uncompressed /Metadata XML stream dict holding the
// given XMP content.
func newMetadataStream(ctx *model.Context, content []byte) (*types.StreamDict, error) {
	d := types.NewDict()
	d.InsertName("Type", "Metadata")
	d.InsertName("Subtype", "XML")

	sd := types.StreamDict{
		Dict:           d,
		Content:        content,
		FilterPipeline: nil, // unfiltered: required for PDF/A /Metadata
	}
	if err := sd.Encode(); err != nil {
		return nil, fmt.Errorf("encode XMP stream: %w", err)
	}
	return &sd, nil
}
