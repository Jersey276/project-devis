# EN 16931 Schematron (CII binding)

Business-rule (Schematron) validation for the Factur-X CII XML, layered on top of
the structural XSD check under [`../schemas/`](../schemas/README.md). XSD proves the
document is *well-formed CII*; this Schematron proves it satisfies the EN 16931
**business rules** (`BR-*`, `BR-CO-*`, `BR-S-*`, `BR-DEC-*` …) — the second gate a
PDP (Plateforme de Dématérialisation Partenaire) applies before accepting an invoice.

## Provenance — vendored verbatim, do not edit

- **Source**: [ConnectingEurope/eInvoicing-EN16931](https://github.com/ConnectingEurope/eInvoicing-EN16931),
  release **`validation-1.3.16`**, asset `en16931-cii-1.3.16.zip`.
- **License**: EUPL 1.2 (the Schematron/XSLT artefacts).

### `en16931/EN16931-CII-validation.xslt`

The **executed** artefact. It is the compiled Schematron, **self-contained** (no
`xsl:include`/`xsl:import`) and written for **XSLT 2.0** (`queryBinding="xslt2"`).
Source path in the release zip: `xslt/EN16931-CII-validation.xslt`.

Because it is XSLT 2.0, **libxml2 / xsltproc (XSLT 1.0) cannot run it** — unlike the
XSD check which uses xmllint. We run it with **Saxon-HE** (see below). Running it on a
CII document yields an SVRL report (`http://purl.oclc.org/dsdl/svrl`); a
`<svrl:failed-assert>` is a rule violation.

### `en16931/sch-source/`

The Schematron **source** (`EN16931-CII-validation.sch` + its `abstract/`, `CII/`,
`codelist/`, `preprocessed/` includes). Kept for provenance and human diffing only —
**not executed**. The `.xslt` above is the compiled form of this tree.

### `saxon/saxon-he-10.9.jar`

[Saxon-HE](https://www.saxonica.com/) 10.9, an XSLT 3.0 processor (MPL 2.0), from
Maven Central (`net.sf.saxon:Saxon-HE:10.9`). See [`saxon/NOTICE.md`](saxon/NOTICE.md).

Saxon-HE **10.x** is used deliberately: it is self-contained, whereas 11+/12+ split
out an `xmlresolver` runtime dependency that a bare `Saxon-HE.jar` no longer bundles
(running 12.x as a single jar fails with `NoClassDefFoundError: org/xmlresolver/Resolver`).
10.9 keeps the vendored toolchain to a single jar.

## How it is run

The conformance harness `../validate_schematron_test.go` shells out to Saxon, mirroring
the xmllint XSD harness: native `java` (JRE 11+) on PATH, else — on a Windows dev box —
`java` inside the **Debian WSL** distro (paths translated with `wslpath`), else the test
**skips**. CI installs a JRE (Temurin 17) for the `export` service so the gate actually
runs there. A dev box with only an old/absent JRE stays green via skip.

## On the absence of a separate "CII-FR" Schematron

There is no standalone FNFE-MPE / Factur-X Schematron to vendor here on top of the one
above. The Factur-X **EN 16931 profile is EN 16931** — the constraints this gate already
enforces. The data once imagined for a `cii-fr/` layer (SIRET, payment means BG-16/IBAN,
statutory mentions) is in fact already emitted by the builder (B2/B4). The genuinely
French layer is not a document Schematron but the **CTC / e-invoicing flow**: directory
routing and the PDP exchange, tracked under B6 (see the invoice compliance roadmap). The
once-anticipated `cii-fr/` directory is therefore intentionally not created.
