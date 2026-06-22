-- Freeze the issuer's and client's ISO 3166-1 alpha-2 country codes at issue
-- time. The export service has no access to the countries table, so the codes
-- must be snapshotted to drive the Factur-X CII buyer/seller country (BT-40 /
-- BT-55) — notably for OSS distance selling where the buyer is in another EU
-- country. '' means "legacy/unknown", in which case the builder falls back to
-- FR (the previous hardcoded behaviour).
ALTER TABLE invoice_party_snapshots
    ADD COLUMN issuer_country_code TEXT NOT NULL DEFAULT '',
    ADD COLUMN client_country_code TEXT NOT NULL DEFAULT '';

ALTER TABLE credit_note_party_snapshots
    ADD COLUMN issuer_country_code TEXT NOT NULL DEFAULT '',
    ADD COLUMN client_country_code TEXT NOT NULL DEFAULT '';
