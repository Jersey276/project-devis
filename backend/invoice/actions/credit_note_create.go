package actions

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"project-devis-invoice/actions/codes"
	invoiceGrpc "project-devis-invoice/services/grpc"
)

func (s *Server) CreateCreditNote(ctx context.Context, req *invoiceGrpc.CreateCreditNoteRequest) (resp *invoiceGrpc.CreateCreditNoteResponse, err error) {
	startedAt := time.Now()
	defer deferObserve("create_credit_note", startedAt, func() (int32, bool) {
		if resp == nil {
			return codes.InternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	if req == nil {
		return &invoiceGrpc.CreateCreditNoteResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	var fieldErrors []*invoiceGrpc.ValidationError
	if strings.TrimSpace(req.UserId) == "" {
		fieldErrors = append(fieldErrors, &invoiceGrpc.ValidationError{Field: "user_id", Message: "Champ requis."})
	}
	if strings.TrimSpace(req.InvoiceId) == "" {
		fieldErrors = append(fieldErrors, &invoiceGrpc.ValidationError{Field: "invoice_id", Message: "Champ requis."})
	}
	if len(fieldErrors) > 0 {
		return &invoiceGrpc.CreateCreditNoteResponse{Success: false, Code: codes.InvalidInput, ValidationErrors: fieldErrors}, nil
	}

	var status string
	var vatExempt bool
	var originInvoiceNumber sql.NullString
	err = s.db.QueryRowContext(ctx,
		`SELECT status, vat_exempt, invoice_number FROM invoices WHERE invoice_id=$1 AND user_id=$2`,
		req.InvoiceId, req.UserId,
	).Scan(&status, &vatExempt, &originInvoiceNumber)
	if err == sql.ErrNoRows {
		return &invoiceGrpc.CreateCreditNoteResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &invoiceGrpc.CreateCreditNoteResponse{Success: false, Code: codes.InternalError}, err
	}
	if status != "ISSUED" && status != "PAID" {
		return &invoiceGrpc.CreateCreditNoteResponse{Success: false, Code: codes.InvoiceNotIssued}, nil
	}

	lineByPos, orderedPos, err := s.loadInvoiceLinesByPosition(ctx, req.InvoiceId)
	if err != nil {
		return &invoiceGrpc.CreateCreditNoteResponse{Success: false, Code: codes.InternalError}, err
	}

	alreadyCredited, err := s.creditedPositions(ctx, req.InvoiceId)
	if err != nil {
		return &invoiceGrpc.CreateCreditNoteResponse{Success: false, Code: codes.InternalError}, err
	}

	selected, isTotal, code := resolveCreditedPositions(req.Positions, orderedPos, alreadyCredited)
	if code != codes.Success {
		return &invoiceGrpc.CreateCreditNoteResponse{Success: false, Code: code}, nil
	}

	computeLines := make([]computeLine, 0, len(selected))
	cnLines := make([]creditNoteLine, 0, len(selected))
	for i, pos := range selected {
		l := lineByPos[pos]
		computeLines = append(computeLines, computeLine{
			ht:        l.lineHTCents,
			taxID:     l.taxID,
			taxRate:   parseRate(l.taxRate),
			taxRateID: l.taxRate,
			taxLabel:  l.taxLabel,
		})
		cnLines = append(cnLines, creditNoteLine{
			position:        int32(i),
			originInvoiceID: req.InvoiceId,
			originPosition:  pos,
			quoteLineID:     l.quoteLineID,
			name:            l.name,
			unit:            l.unit,
			quantity:        l.quantity,
			unitPriceCents:  l.unitPriceCents,
			lineHTCents:     l.lineHTCents,
			taxID:           l.taxID,
			taxRate:         l.taxRate,
			taxLabel:        l.taxLabel,
		})
	}

	totals := computeTotals(computeLines, vatExempt)

	party, err := s.loadInvoicePartySnapshot(ctx, req.InvoiceId)
	if err != nil {
		return &invoiceGrpc.CreateCreditNoteResponse{Success: false, Code: codes.InternalError}, err
	}

	issuedAt := time.Now().In(invoiceTZ)
	year := issuedAt.Year()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return &invoiceGrpc.CreateCreditNoteResponse{Success: false, Code: codes.InternalError}, err
	}
	defer tx.Rollback()

	var lockedStatus string
	if err := tx.QueryRowContext(ctx,
		`SELECT status FROM invoices WHERE invoice_id=$1 FOR UPDATE`, req.InvoiceId,
	).Scan(&lockedStatus); err != nil {
		return &invoiceGrpc.CreateCreditNoteResponse{Success: false, Code: codes.InternalError}, err
	}
	if lockedStatus != "ISSUED" && lockedStatus != "PAID" {
		return &invoiceGrpc.CreateCreditNoteResponse{Success: false, Code: codes.InvoiceNotIssued}, nil
	}

	number, seq, err := allocateCreditNoteNumber(ctx, tx, req.UserId, year)
	if err != nil {
		return &invoiceGrpc.CreateCreditNoteResponse{Success: false, Code: codes.InternalError}, err
	}

	creditNoteID := uuid.New().String()
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO credit_notes (
			credit_note_id, user_id, invoice_id, credit_note_number, number_year, number_seq,
			is_total, reason, issued_at,
			total_ht_cents, total_vat_cents, total_ttc_cents, vat_exempt
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		creditNoteID, req.UserId, req.InvoiceId, number, year, seq,
		isTotal, strings.TrimSpace(req.Reason), issuedAt,
		totals.totalHT, totals.totalVAT, totals.totalTTC, vatExempt,
	); err != nil {
		return &invoiceGrpc.CreateCreditNoteResponse{Success: false, Code: codes.InternalError}, err
	}

	if err := writeCreditNoteSnapshots(ctx, tx, creditNoteID, party, cnLines, totals.breakdown); err != nil {
		if isUniqueViolation(err, "credit_note_lines_origin_unique") {

			return &invoiceGrpc.CreateCreditNoteResponse{Success: false, Code: codes.CreditNoteLineAlreadyCredited}, nil
		}
		return &invoiceGrpc.CreateCreditNoteResponse{Success: false, Code: codes.InternalError}, err
	}

	contentHash := computeContentHash(sealableDoc{
		userID:              req.UserId,
		docType:             "CREDIT_NOTE",
		number:              number,
		issuedAt:            issuedAt,
		totalHT:             totals.totalHT,
		totalVAT:            totals.totalVAT,
		totalTTC:            totals.totalTTC,
		vatExempt:           vatExempt,
		originInvoiceNumber: originInvoiceNumber.String,
		lines:               sealLinesFromCreditNoteLines(cnLines),
	})
	if _, _, err := sealDocument(ctx, tx, req.UserId, "CREDIT_NOTE", creditNoteID, contentHash); err != nil {
		return &invoiceGrpc.CreateCreditNoteResponse{Success: false, Code: codes.SealError}, err
	}

	if err := tx.Commit(); err != nil {
		return &invoiceGrpc.CreateCreditNoteResponse{Success: false, Code: codes.InternalError}, err
	}

	return &invoiceGrpc.CreateCreditNoteResponse{Success: true, Code: codes.Success, CreditNoteId: creditNoteID, CreditNoteNumber: number}, nil
}

func resolveCreditedPositions(requested []int32, allPositions []int32, alreadyCredited map[int32]struct{}) (selected []int32, isTotal bool, code int32) {
	allSet := make(map[int32]struct{}, len(allPositions))
	for _, p := range allPositions {
		allSet[p] = struct{}{}
	}

	if len(requested) == 0 {

		for _, p := range allPositions {
			if _, done := alreadyCredited[p]; !done {
				selected = append(selected, p)
			}
		}
		if len(selected) == 0 {
			return nil, false, codes.CreditNoteNoLinesLeft
		}
		return selected, true, codes.Success
	}

	seen := make(map[int32]struct{}, len(requested))
	for _, p := range requested {
		if _, ok := allSet[p]; !ok {
			return nil, false, codes.InvalidInput
		}
		if _, dup := seen[p]; dup {
			return nil, false, codes.InvalidInput
		}
		seen[p] = struct{}{}
		if _, done := alreadyCredited[p]; done {
			return nil, false, codes.CreditNoteLineAlreadyCredited
		}
		selected = append(selected, p)
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i] < selected[j] })

	return selected, len(selected) == len(allPositions), codes.Success
}

func (s *Server) loadInvoiceLinesByPosition(ctx context.Context, invoiceID string) (map[int32]lineSnapshot, []int32, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT position, quote_line_id, name, unit, quantity, unit_price_cents, line_ht_cents, tax_id, tax_rate, tax_label
		 FROM invoice_line_snapshots WHERE invoice_id=$1 ORDER BY position`,
		invoiceID,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	byPos := make(map[int32]lineSnapshot)
	var ordered []int32
	for rows.Next() {
		var l lineSnapshot
		if err := rows.Scan(&l.position, &l.quoteLineID, &l.name, &l.unit, &l.quantity,
			&l.unitPriceCents, &l.lineHTCents, &l.taxID, &l.taxRate, &l.taxLabel); err != nil {
			return nil, nil, err
		}
		byPos[l.position] = l
		ordered = append(ordered, l.position)
	}
	return byPos, ordered, rows.Err()
}

func (s *Server) creditedPositions(ctx context.Context, invoiceID string) (map[int32]struct{}, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT origin_position FROM credit_note_lines WHERE origin_invoice_id=$1`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	set := make(map[int32]struct{})
	for rows.Next() {
		var p int32
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		set[p] = struct{}{}
	}
	return set, rows.Err()
}

func (s *Server) loadInvoicePartySnapshot(ctx context.Context, invoiceID string) (partySnapshot, error) {
	var p partySnapshot
	err := s.db.QueryRowContext(ctx,
		`SELECT issuer_company, issuer_siren, issuer_vat, issuer_email, issuer_phone,
		        issuer_street, issuer_additional, issuer_zip, issuer_city,
		        client_first_name, client_last_name, client_company, client_email,
		        client_street, client_additional, client_zip, client_city, client_type, client_country_id, oss_applied,
		        issuer_country_code, client_country_code, counts_toward_oss_threshold, issuer_iban, issuer_bic,
		        issuer_siret, client_siret
		 FROM invoice_party_snapshots WHERE invoice_id=$1`,
		invoiceID,
	).Scan(
		&p.issuerCompany, &p.issuerSiren, &p.issuerVat, &p.issuerEmail, &p.issuerPhone,
		&p.issuerStreet, &p.issuerAdditional, &p.issuerZip, &p.issuerCity,
		&p.clientFirstName, &p.clientLastName, &p.clientCompany, &p.clientEmail,
		&p.clientStreet, &p.clientAdditional, &p.clientZip, &p.clientCity, &p.clientType, &p.clientCountryID, &p.ossApplied,
		&p.issuerCountryCode, &p.clientCountryCode, &p.countsTowardThreshold, &p.issuerIban, &p.issuerBic,
		&p.issuerSiret, &p.clientSiret,
	)
	return p, err
}

func isUniqueViolation(err error, constraint string) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505" && pqErr.Constraint == constraint
	}
	return false
}
