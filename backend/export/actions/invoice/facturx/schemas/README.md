# CII XSD schemas (EN 16931 / Factur-X)

Official UN/CEFACT Cross Industry Invoice (CII) XSD schemas, SCRDM **D16B**
subset — the structure a Factur-X EN 16931 document is validated against.

- **Source**: [ConnectingEurope/eInvoicing-EN16931](https://github.com/ConnectingEurope/eInvoicing-EN16931),
  path `cii/schema/D16B SCRDM (Subset)/coupled clm/CII/uncefact`.
- **Root schema**: `uncefact/data/standard/CrossIndustryInvoice_100pD16B.xsd`.
  It imports the Reusable/Qualified/Unqualified data-type schemas, which in turn
  import the UN/ECE and ISO code lists under `uncefact/codelist` and
  `uncefact/identifierlist`. All 54 files must stay together for resolution.

These files are **vendored verbatim** — do not edit. They back the XSD
conformance test (`validate_xsd_test.go`), which proves the generator emits
structurally valid CII. The test runs `xmllint` and skips when it is absent
(e.g. a Windows dev box without libxml2-utils); CI must install `libxml2-utils`.
