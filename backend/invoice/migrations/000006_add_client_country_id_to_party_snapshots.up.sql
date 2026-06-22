-- Freeze the client's country id on the invoice and credit-note party snapshots.
-- Drives OSS destination-country VAT at issue time. 0 means "legacy snapshot,
-- country unknown", matching the NOT NULL DEFAULT convention of the other
-- party-snapshot columns.
ALTER TABLE invoice_party_snapshots
    ADD COLUMN client_country_id INT NOT NULL DEFAULT 0;

ALTER TABLE credit_note_party_snapshots
    ADD COLUMN client_country_id INT NOT NULL DEFAULT 0;
