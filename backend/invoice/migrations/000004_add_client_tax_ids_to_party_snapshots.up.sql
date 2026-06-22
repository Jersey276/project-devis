-- The buyer SIREN/VAT are required by the EN 16931 / Factur-X structured format
-- but were not captured in the party snapshot. Add them additively; existing
-- rows default to '' (the XML omits the buyer tax ids when empty).
ALTER TABLE invoice_party_snapshots
    ADD COLUMN client_siren TEXT NOT NULL DEFAULT '',
    ADD COLUMN client_vat   TEXT NOT NULL DEFAULT '';
