package actions

import (
	"context"

	invoiceGrpc "project-devis-invoice/services/grpc"
)

func (s *Server) loadSnapshot(ctx context.Context, invoiceID string, details *invoiceGrpc.InvoiceDetails) error {
	var p partySnapshot
	err := s.db.QueryRowContext(ctx,
		`SELECT issuer_company, issuer_siren, issuer_vat, issuer_email, issuer_phone,
		        issuer_street, issuer_additional, issuer_zip, issuer_city,
		        client_first_name, client_last_name, client_company, client_siren, client_vat, client_email,
		        client_street, client_additional, client_zip, client_city, client_type, client_country_id, oss_applied,
		        issuer_country_code, client_country_code, issuer_iban, issuer_bic,
		        issuer_siret, client_siret
		 FROM invoice_party_snapshots WHERE invoice_id=$1`,
		invoiceID,
	).Scan(
		&p.issuerCompany, &p.issuerSiren, &p.issuerVat, &p.issuerEmail, &p.issuerPhone,
		&p.issuerStreet, &p.issuerAdditional, &p.issuerZip, &p.issuerCity,
		&p.clientFirstName, &p.clientLastName, &p.clientCompany, &p.clientSiren, &p.clientVat, &p.clientEmail,
		&p.clientStreet, &p.clientAdditional, &p.clientZip, &p.clientCity, &p.clientType, &p.clientCountryID, &p.ossApplied,
		&p.issuerCountryCode, &p.clientCountryCode, &p.issuerIban, &p.issuerBic,
		&p.issuerSiret, &p.clientSiret,
	)
	if err != nil {
		return err
	}
	details.Issuer = partyToProto(p, true)
	details.Client = partyToProto(p, false)
	details.OssApplied = p.ossApplied

	lineRows, err := s.db.QueryContext(ctx,
		`SELECT quote_line_id, name, unit, quantity, unit_price_cents, line_ht_cents, tax_id, tax_rate, tax_label
		 FROM invoice_line_snapshots WHERE invoice_id=$1 ORDER BY position`,
		invoiceID,
	)
	if err != nil {
		return err
	}
	defer lineRows.Close()
	for lineRows.Next() {
		l := &invoiceGrpc.InvoiceLine{}
		if err := lineRows.Scan(&l.QuoteLineId, &l.Name, &l.Unit, &l.Quantity,
			&l.UnitPriceCents, &l.LineHtCents, &l.TaxId, &l.TaxRate, &l.TaxLabel); err != nil {
			return err
		}
		details.Lines = append(details.Lines, l)
	}
	if err := lineRows.Err(); err != nil {
		return err
	}

	vatRows, err := s.db.QueryContext(ctx,
		`SELECT tax_rate, base_ht_cents, vat_cents FROM invoice_vat_breakdown_snapshots
		 WHERE invoice_id=$1 ORDER BY (tax_rate)::numeric`,
		invoiceID,
	)
	if err != nil {
		return err
	}
	defer vatRows.Close()
	for vatRows.Next() {
		v := &invoiceGrpc.InvoiceVatLine{}
		if err := vatRows.Scan(&v.TaxRate, &v.BaseHtCents, &v.VatCents); err != nil {
			return err
		}
		details.VatBreakdown = append(details.VatBreakdown, v)
	}
	return vatRows.Err()
}

func partyToProto(p partySnapshot, issuer bool) *invoiceGrpc.InvoiceParty {
	if issuer {
		return &invoiceGrpc.InvoiceParty{
			Company:          p.issuerCompany,
			Siren:            p.issuerSiren,
			Siret:            p.issuerSiret,
			Vat:              p.issuerVat,
			Email:            p.issuerEmail,
			Phone:            p.issuerPhone,
			Street:           p.issuerStreet,
			AdditionalStreet: p.issuerAdditional,
			ZipCode:          p.issuerZip,
			City:             p.issuerCity,
			CountryCode:      p.issuerCountryCode,
			Iban:             p.issuerIban,
			Bic:              p.issuerBic,
		}
	}
	return &invoiceGrpc.InvoiceParty{
		Company:          p.clientCompany,
		FirstName:        p.clientFirstName,
		LastName:         p.clientLastName,
		Siren:            p.clientSiren,
		Siret:            p.clientSiret,
		Vat:              p.clientVat,
		Email:            p.clientEmail,
		Street:           p.clientStreet,
		AdditionalStreet: p.clientAdditional,
		ZipCode:          p.clientZip,
		City:             p.clientCity,
		ClientType:       p.clientType,
		ClientCountryId:  p.clientCountryID,
		CountryCode:      p.clientCountryCode,
	}
}
