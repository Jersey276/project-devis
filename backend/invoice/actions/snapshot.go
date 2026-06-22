package actions

import (
	"context"
	"database/sql"
)

func writeSnapshots(ctx context.Context, tx *sql.Tx, invoiceID string, r *resolvedInvoice, breakdown []vatBucket) error {
	p := r.parties
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO invoice_party_snapshots (
			invoice_id,
			issuer_company, issuer_siren, issuer_vat, issuer_email, issuer_phone, issuer_logo_url,
			issuer_street, issuer_additional, issuer_zip, issuer_city,
			client_first_name, client_last_name, client_company, client_siren, client_vat, client_email,
			client_street, client_additional, client_zip, client_city, client_type, client_country_id, oss_applied,
			issuer_country_code, client_country_code, counts_toward_oss_threshold,
			issuer_iban, issuer_bic,
			issuer_siret, client_siret
		) VALUES (
			$1,
			$2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22, $23, $24,
			$25, $26, $27,
			$28, $29,
			$30, $31
		)`,
		invoiceID,
		p.issuerCompany, p.issuerSiren, p.issuerVat, p.issuerEmail, p.issuerPhone, p.issuerLogoURL,
		p.issuerStreet, p.issuerAdditional, p.issuerZip, p.issuerCity,
		p.clientFirstName, p.clientLastName, p.clientCompany, p.clientSiren, p.clientVat, p.clientEmail,
		p.clientStreet, p.clientAdditional, p.clientZip, p.clientCity, p.clientType, p.clientCountryID, p.ossApplied,
		p.issuerCountryCode, p.clientCountryCode, p.countsTowardThreshold,
		p.issuerIban, p.issuerBic,
		p.issuerSiret, p.clientSiret,
	); err != nil {
		return err
	}

	for _, l := range r.lines {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO invoice_line_snapshots (
				invoice_id, position, quote_line_id, name, unit, quantity,
				unit_price_cents, line_ht_cents, tax_id, tax_rate, tax_label
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			invoiceID, l.position, l.quoteLineID, l.name, l.unit, l.quantity,
			l.unitPriceCents, l.lineHTCents, l.taxID, l.taxRate, l.taxLabel,
		); err != nil {
			return err
		}
	}

	for _, b := range breakdown {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO invoice_vat_breakdown_snapshots (invoice_id, tax_rate, base_ht_cents, vat_cents)
			 VALUES ($1, $2, $3, $4)`,
			invoiceID, b.rate, b.baseHT, b.vat,
		); err != nil {
			return err
		}
	}

	return nil
}
