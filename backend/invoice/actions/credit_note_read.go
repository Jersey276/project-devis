package actions

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"project-devis-invoice/actions/codes"
	invoiceGrpc "project-devis-invoice/services/grpc"
)

func (s *Server) GetCreditNote(ctx context.Context, req *invoiceGrpc.GetCreditNoteRequest) (resp *invoiceGrpc.GetCreditNoteResponse, err error) {
	startedAt := time.Now()
	defer deferObserve("get_credit_note", startedAt, func() (int32, bool) {
		if resp == nil {
			return codes.InternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	if req == nil || strings.TrimSpace(req.CreditNoteId) == "" || strings.TrimSpace(req.UserId) == "" {
		return &invoiceGrpc.GetCreditNoteResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var (
		invoiceID     string
		invoiceNumber sql.NullString
		number        string
		issuedAt      sql.NullTime
		reason        string
		isTotal       bool
		totalHT       int64
		totalVAT      int64
		totalTTC      int64
		vatExempt     bool
	)
	err = s.db.QueryRowContext(ctx,
		`SELECT cn.invoice_id, i.invoice_number, cn.credit_note_number, cn.issued_at, cn.reason,
		        cn.is_total, cn.total_ht_cents, cn.total_vat_cents, cn.total_ttc_cents, cn.vat_exempt
		 FROM credit_notes cn
		 JOIN invoices i ON i.invoice_id = cn.invoice_id
		 WHERE cn.credit_note_id=$1 AND cn.user_id=$2`,
		req.CreditNoteId, req.UserId,
	).Scan(&invoiceID, &invoiceNumber, &number, &issuedAt, &reason,
		&isTotal, &totalHT, &totalVAT, &totalTTC, &vatExempt)
	if err == sql.ErrNoRows {
		return &invoiceGrpc.GetCreditNoteResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &invoiceGrpc.GetCreditNoteResponse{Success: false, Code: codes.InternalError}, err
	}

	details := &invoiceGrpc.CreditNoteDetails{
		CreditNoteId:     req.CreditNoteId,
		UserId:           req.UserId,
		InvoiceId:        invoiceID,
		InvoiceNumber:    invoiceNumber.String,
		CreditNoteNumber: number,
		IssuedAt:         formatNullTime(issuedAt, time.RFC3339),
		Reason:           reason,
		IsTotal:          isTotal,
		TotalHtCents:     totalHT,
		TotalVatCents:    totalVAT,
		TotalTtcCents:    totalTTC,
		VatExempt:        vatExempt,
	}

	if err := s.loadCreditNoteSnapshot(ctx, req.CreditNoteId, details); err != nil {
		return &invoiceGrpc.GetCreditNoteResponse{Success: false, Code: codes.InternalError}, err
	}
	return &invoiceGrpc.GetCreditNoteResponse{Success: true, Code: codes.Success, CreditNote: details}, nil
}

func (s *Server) loadCreditNoteSnapshot(ctx context.Context, creditNoteID string, details *invoiceGrpc.CreditNoteDetails) error {
	var p partySnapshot
	err := s.db.QueryRowContext(ctx,
		`SELECT issuer_company, issuer_siren, issuer_vat, issuer_email, issuer_phone, issuer_logo_url,
		        issuer_street, issuer_additional, issuer_zip, issuer_city,
		        client_first_name, client_last_name, client_company, client_email,
		        client_street, client_additional, client_zip, client_city, client_type, client_country_id, oss_applied,
		        issuer_country_code, client_country_code
		 FROM credit_note_party_snapshots WHERE credit_note_id=$1`,
		creditNoteID,
	).Scan(
		&p.issuerCompany, &p.issuerSiren, &p.issuerVat, &p.issuerEmail, &p.issuerPhone, &p.issuerLogoURL,
		&p.issuerStreet, &p.issuerAdditional, &p.issuerZip, &p.issuerCity,
		&p.clientFirstName, &p.clientLastName, &p.clientCompany, &p.clientEmail,
		&p.clientStreet, &p.clientAdditional, &p.clientZip, &p.clientCity, &p.clientType, &p.clientCountryID, &p.ossApplied,
		&p.issuerCountryCode, &p.clientCountryCode,
	)
	if err != nil {
		return err
	}
	details.Issuer = partyToProto(p, true)
	details.Client = partyToProto(p, false)
	details.OssApplied = p.ossApplied

	lineRows, err := s.db.QueryContext(ctx,
		`SELECT quote_line_id, name, unit, quantity, unit_price_cents, line_ht_cents, tax_id, tax_rate, tax_label
		 FROM credit_note_lines WHERE credit_note_id=$1 ORDER BY position`,
		creditNoteID,
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
		`SELECT tax_rate, base_ht_cents, vat_cents FROM credit_note_vat_breakdown_snapshots
		 WHERE credit_note_id=$1 ORDER BY (tax_rate)::numeric`,
		creditNoteID,
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

func (s *Server) ListCreditNotes(ctx context.Context, req *invoiceGrpc.ListCreditNotesRequest) (resp *invoiceGrpc.ListCreditNotesResponse, err error) {
	startedAt := time.Now()
	defer deferObserve("list_credit_notes", startedAt, func() (int32, bool) {
		if resp == nil {
			return codes.InternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	if req == nil || strings.TrimSpace(req.UserId) == "" {
		return &invoiceGrpc.ListCreditNotesResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	query := `SELECT cn.credit_note_id, cn.credit_note_number, cn.invoice_id, i.invoice_number,
	                 cn.issued_at, cn.is_total, cn.total_ttc_cents
	          FROM credit_notes cn
	          JOIN invoices i ON i.invoice_id = cn.invoice_id
	          WHERE cn.user_id=$1`
	args := []any{req.UserId}
	if inv := strings.TrimSpace(req.InvoiceId); inv != "" {
		query += ` AND cn.invoice_id=$2`
		args = append(args, inv)
	}
	query += ` ORDER BY cn.created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return &invoiceGrpc.ListCreditNotesResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	out := make([]*invoiceGrpc.CreditNoteSummary, 0)
	for rows.Next() {
		var (
			id, number, invoiceID string
			invoiceNumber         sql.NullString
			issuedAt              sql.NullTime
			isTotal               bool
			totalTTC              int64
		)
		if err := rows.Scan(&id, &number, &invoiceID, &invoiceNumber, &issuedAt, &isTotal, &totalTTC); err != nil {
			return &invoiceGrpc.ListCreditNotesResponse{Success: false, Code: codes.InternalError}, err
		}
		out = append(out, &invoiceGrpc.CreditNoteSummary{
			CreditNoteId:     id,
			CreditNoteNumber: number,
			InvoiceId:        invoiceID,
			InvoiceNumber:    invoiceNumber.String,
			IssuedAt:         formatNullTime(issuedAt, time.RFC3339),
			IsTotal:          isTotal,
			TotalTtcCents:    totalTTC,
		})
	}
	if err := rows.Err(); err != nil {
		return &invoiceGrpc.ListCreditNotesResponse{Success: false, Code: codes.InternalError}, err
	}

	return &invoiceGrpc.ListCreditNotesResponse{Success: true, Code: codes.Success, CreditNotes: out}, nil
}
