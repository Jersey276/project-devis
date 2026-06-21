# Saxon-HE 10.9

- **Artefact**: `saxon-he-10.9.jar`
- **Coordinates**: `net.sf.saxon:Saxon-HE:10.9` (Maven Central)
- **SHA-256**: `491d8edf4ec811d15c2b2417b007218b9b938f15e4dfbad004025beb4e70e960`
- **Vendor**: Saxonica Limited — <https://www.saxonica.com/>
- **License**: Mozilla Public License 2.0 (MPL-2.0) — full text in [`MPL-2.0.txt`](MPL-2.0.txt),
  canonical copy at <https://www.mozilla.org/MPL/2.0/>
- **Main-Class**: `net.sf.saxon.Transform` (runs via `java -jar saxon-he-10.9.jar`)

Vendored verbatim and unmodified. Used only at test time to run the EN 16931 CII
Schematron XSLT (XSLT 2.0) and produce an SVRL report. Saxon-HE **10.x** is chosen
because it is self-contained; 11+/12+ require an external `xmlresolver` dependency.
