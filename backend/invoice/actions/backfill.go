package actions

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

func BackfillSeals(ctx context.Context, db *sql.DB) error {
	users, err := usersWithUnsealedDocs(ctx, db)
	if err != nil {
		return fmt.Errorf("list users to backfill: %w", err)
	}
	total := 0
	for _, userID := range users {
		n, err := backfillUser(ctx, db, userID)
		if err != nil {
			return fmt.Errorf("backfill user %s: %w", userID, err)
		}
		total += n
	}
	if total > 0 {
		log.Printf("seal backfill: sealed %d legacy document(s) across %d issuer(s)", total, len(users))
	}
	return nil
}

func usersWithUnsealedDocs(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT user_id FROM (
			SELECT i.user_id
			FROM invoices i
			LEFT JOIN document_seals ds ON ds.doc_type='INVOICE' AND ds.doc_id=i.invoice_id
			WHERE i.status IN ('ISSUED','PAID','CANCELLED') AND ds.doc_id IS NULL
			UNION
			SELECT cn.user_id
			FROM credit_notes cn
			LEFT JOIN document_seals ds ON ds.doc_type='CREDIT_NOTE' AND ds.doc_id=cn.credit_note_id
			WHERE ds.doc_id IS NULL
		) u
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

type legacyDoc struct {
	docType             string
	docID               string
	number              string
	issuedAt            sql.NullTime
	totalHT             int64
	totalVAT            int64
	totalTTC            int64
	vatExempt           bool
	originInvoiceNumber string
}

func backfillUser(ctx context.Context, db *sql.DB, userID string) (int, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT doc_type, doc_id, number, issued_at, total_ht_cents, total_vat_cents,
		       total_ttc_cents, vat_exempt, origin_invoice_number, number_seq
		FROM (
			SELECT 'INVOICE' AS doc_type, i.invoice_id AS doc_id, i.invoice_number AS number,
			       i.issued_at, i.total_ht_cents, i.total_vat_cents, i.total_ttc_cents,
			       i.vat_exempt, '' AS origin_invoice_number, i.number_seq
			FROM invoices i
			LEFT JOIN document_seals ds ON ds.doc_type='INVOICE' AND ds.doc_id=i.invoice_id
			WHERE i.user_id=$1 AND i.status IN ('ISSUED','PAID','CANCELLED') AND ds.doc_id IS NULL
			UNION ALL
			SELECT 'CREDIT_NOTE', cn.credit_note_id, cn.credit_note_number,
			       cn.issued_at, cn.total_ht_cents, cn.total_vat_cents, cn.total_ttc_cents,
			       cn.vat_exempt, COALESCE(oi.invoice_number, ''), cn.number_seq
			FROM credit_notes cn
			JOIN invoices oi ON oi.invoice_id = cn.invoice_id
			LEFT JOIN document_seals ds ON ds.doc_type='CREDIT_NOTE' AND ds.doc_id=cn.credit_note_id
			WHERE cn.user_id=$1 AND ds.doc_id IS NULL
		) docs
		ORDER BY issued_at ASC, doc_type ASC, number_seq ASC
	`, userID)
	if err != nil {
		return 0, err
	}

	var docs []legacyDoc
	for rows.Next() {
		var d legacyDoc
		var seq sql.NullInt64
		if err := rows.Scan(&d.docType, &d.docID, &d.number, &d.issuedAt,
			&d.totalHT, &d.totalVAT, &d.totalTTC, &d.vatExempt, &d.originInvoiceNumber, &seq); err != nil {
			rows.Close()
			return 0, err
		}
		docs = append(docs, d)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}

	for _, d := range docs {
		lines, err := loadSealLinesForDoc(ctx, tx, d.docType, d.docID)
		if err != nil {
			return 0, err
		}
		contentHash := computeContentHash(sealableDoc{
			userID:              userID,
			docType:             d.docType,
			number:              d.number,
			issuedAt:            d.issuedAt.Time,
			totalHT:             d.totalHT,
			totalVAT:            d.totalVAT,
			totalTTC:            d.totalTTC,
			vatExempt:           d.vatExempt,
			originInvoiceNumber: d.originInvoiceNumber,
			lines:               lines,
		})
		if _, _, err := sealDocument(ctx, tx, userID, d.docType, d.docID, contentHash); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return len(docs), nil
}

func loadSealLinesForDoc(ctx context.Context, tx *sql.Tx, docType, docID string) ([]sealLine, error) {
	var query string
	switch docType {
	case "INVOICE":
		query = `SELECT name, quantity, unit, unit_price_cents, line_ht_cents, tax_rate, tax_label
		         FROM invoice_line_snapshots WHERE invoice_id=$1 ORDER BY position`
	case "CREDIT_NOTE":
		query = `SELECT name, quantity, unit, unit_price_cents, line_ht_cents, tax_rate, tax_label
		         FROM credit_note_lines WHERE credit_note_id=$1 ORDER BY position`
	default:
		return nil, fmt.Errorf("unknown doc type %q", docType)
	}
	rows, err := tx.QueryContext(ctx, query, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lines []sealLine
	for rows.Next() {
		var l sealLine
		if err := rows.Scan(&l.name, &l.quantity, &l.unit, &l.unitPriceCents, &l.lineHTCents, &l.taxRate, &l.taxLabel); err != nil {
			return nil, err
		}
		lines = append(lines, l)
	}
	return lines, rows.Err()
}
