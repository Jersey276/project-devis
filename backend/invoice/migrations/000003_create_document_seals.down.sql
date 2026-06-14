DROP TRIGGER IF EXISTS credit_note_vat_breakdown_snapshots_immutable ON credit_note_vat_breakdown_snapshots;
DROP TRIGGER IF EXISTS credit_note_lines_immutable ON credit_note_lines;
DROP TRIGGER IF EXISTS credit_note_party_snapshots_immutable ON credit_note_party_snapshots;
DROP TRIGGER IF EXISTS invoice_vat_breakdown_snapshots_immutable ON invoice_vat_breakdown_snapshots;
DROP TRIGGER IF EXISTS invoice_line_snapshots_immutable ON invoice_line_snapshots;
DROP TRIGGER IF EXISTS invoice_party_snapshots_immutable ON invoice_party_snapshots;
DROP TRIGGER IF EXISTS credit_notes_immutable ON credit_notes;
DROP TRIGGER IF EXISTS invoices_immutable ON invoices;
DROP TRIGGER IF EXISTS document_seals_no_update_delete ON document_seals;

DROP FUNCTION IF EXISTS trg_snapshot_immutable();
DROP FUNCTION IF EXISTS trg_credit_notes_immutable();
DROP FUNCTION IF EXISTS trg_invoices_immutable();
DROP FUNCTION IF EXISTS trg_document_seals_immutable();

DROP TABLE IF EXISTS chain_heads;
DROP TABLE IF EXISTS document_seals;
