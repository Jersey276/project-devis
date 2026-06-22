-- OSS distance-selling threshold assiette flag, frozen at issue time.
-- Rationale: docs/adr/0002-oss-seuil-bascule-automatique.md.
-- FALSE matches legacy rows (domestic / B2B / pre-feature), so NOT NULL is safe.
ALTER TABLE invoice_party_snapshots
    ADD COLUMN counts_toward_oss_threshold BOOLEAN NOT NULL DEFAULT FALSE;

-- Mirrored on credit notes; not populated yet (avoirs excluded from the assiette).
ALTER TABLE credit_note_party_snapshots
    ADD COLUMN counts_toward_oss_threshold BOOLEAN NOT NULL DEFAULT FALSE;

-- Covers the yearly cumulative query (join to the snapshot is by PK).
CREATE INDEX invoices_oss_cumulative_idx
    ON invoices (user_id, issued_at)
    WHERE status IN ('ISSUED', 'PAID');
