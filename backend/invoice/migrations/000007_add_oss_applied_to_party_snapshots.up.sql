-- Freeze whether the OSS distance-selling regime (destination-country VAT) was
-- applied at issue time. This is the only reliable source for the PDF's OSS
-- mention years later, since the seller's oss_enabled flag is mutable. FALSE
-- means "domestic billing or legacy snapshot", matching the NOT NULL DEFAULT
-- convention of the other party-snapshot columns.
ALTER TABLE invoice_party_snapshots
    ADD COLUMN oss_applied BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE credit_note_party_snapshots
    ADD COLUMN oss_applied BOOLEAN NOT NULL DEFAULT FALSE;
