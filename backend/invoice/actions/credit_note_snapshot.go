package actions

import (
	"context"
	"database/sql"
)

// creditNoteLine is one frozen credited line: the snapshot of an invoice line
// plus the (origin_invoice_id, origin_position) it credits — the latter feeds
// the over-crediting UNIQUE constraint.
type creditNoteLine struct {
	position        int32
	originInvoiceID string
	originPosition  int32
	quoteLineID     string
	name            string
	unit            string
	quantity        string
	unitPriceCents  int64
	lineHTCents     int64
	taxID           int32
	taxRate         string
	taxLabel        string
}

// writeCreditNoteSnapshots persists the frozen party block, credited lines and
// VAT breakdown inside the creation transaction. Inserting a line whose
// (origin_invoice_id, origin_position) is already credited raises a unique
// violation (23505) — the caller maps it to CreditNoteLineAlreadyCredited.
func writeCreditNoteSnapshots(ctx context.Context, tx *sql.Tx, creditNoteID string, p partySnapshot, lines []creditNoteLine, breakdown []vatBucket) error {
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO credit_note_party_snapshots (
			credit_note_id,
			issuer_company, issuer_siren, issuer_vat, issuer_email, issuer_phone, issuer_logo_url,
			issuer_street, issuer_additional, issuer_zip, issuer_city,
			client_first_name, client_last_name, client_company, client_email,
			client_street, client_additional, client_zip, client_city, client_type, client_country_id, oss_applied,
			issuer_country_code, client_country_code
		) VALUES (
			$1,
			$2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11,
			$12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22,
			$23, $24
		)`,
		creditNoteID,
		p.issuerCompany, p.issuerSiren, p.issuerVat, p.issuerEmail, p.issuerPhone, p.issuerLogoURL,
		p.issuerStreet, p.issuerAdditional, p.issuerZip, p.issuerCity,
		p.clientFirstName, p.clientLastName, p.clientCompany, p.clientEmail,
		p.clientStreet, p.clientAdditional, p.clientZip, p.clientCity, p.clientType, p.clientCountryID, p.ossApplied,
		p.issuerCountryCode, p.clientCountryCode,
	); err != nil {
		return err
	}

	for _, l := range lines {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO credit_note_lines (
				credit_note_id, position, origin_invoice_id, origin_position,
				quote_line_id, name, unit, quantity,
				unit_price_cents, line_ht_cents, tax_id, tax_rate, tax_label
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
			creditNoteID, l.position, l.originInvoiceID, l.originPosition,
			l.quoteLineID, l.name, l.unit, l.quantity,
			l.unitPriceCents, l.lineHTCents, l.taxID, l.taxRate, l.taxLabel,
		); err != nil {
			return err
		}
	}

	for _, b := range breakdown {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO credit_note_vat_breakdown_snapshots (credit_note_id, tax_rate, base_ht_cents, vat_cents)
			 VALUES ($1, $2, $3, $4)`,
			creditNoteID, b.rate, b.baseHT, b.vat,
		); err != nil {
			return err
		}
	}

	return nil
}
