package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"project-devis-invoice/actions/codes"
	invoiceGrpc "project-devis-invoice/services/grpc"
)

func (s *Server) VerifyChain(ctx context.Context, req *invoiceGrpc.VerifyChainRequest) (resp *invoiceGrpc.VerifyChainResponse, err error) {
	startedAt := time.Now()
	defer deferObserve("verify_chain", startedAt, func() (int32, bool) {
		if resp == nil {
			return codes.InternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	if req == nil || strings.TrimSpace(req.UserId) == "" {
		return &invoiceGrpc.VerifyChainResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT doc_type, doc_id, chain_index, content_hash, prev_hash, chain_hash
		 FROM document_seals WHERE user_id=$1 ORDER BY chain_index`,
		req.UserId,
	)
	if err != nil {
		return &invoiceGrpc.VerifyChainResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	prevHash := genesisHash
	var expectedIndex int64
	var checked int64

	for rows.Next() {
		var docType, docID, contentHash, storedPrev, chainHash string
		var index int64
		if err := rows.Scan(&docType, &docID, &index, &contentHash, &storedPrev, &chainHash); err != nil {
			return &invoiceGrpc.VerifyChainResponse{Success: false, Code: codes.InternalError}, err
		}

		if index != expectedIndex {
			return broken(docType, docID, index, fmt.Sprintf("index non contigu : attendu %d, obtenu %d", expectedIndex, index)), nil
		}

		if storedPrev != prevHash {
			return broken(docType, docID, index, "prev_hash ne correspond pas au maillon précédent"), nil
		}

		if computeChainHash(storedPrev, contentHash, index) != chainHash {
			return broken(docType, docID, index, "chain_hash altéré"), nil
		}

		recomputed, code, err := s.recomputeContentHash(ctx, req.UserId, docType, docID)
		if err != nil {
			return &invoiceGrpc.VerifyChainResponse{Success: false, Code: code}, err
		}
		if recomputed != contentHash {
			return broken(docType, docID, index, "le contenu du document a été modifié"), nil
		}

		prevHash = chainHash
		expectedIndex++
		checked++
	}
	if err := rows.Err(); err != nil {
		return &invoiceGrpc.VerifyChainResponse{Success: false, Code: codes.InternalError}, err
	}

	return &invoiceGrpc.VerifyChainResponse{Success: true, Code: codes.Success, Ok: true, Checked: checked}, nil
}

func broken(docType, docID string, index int64, reason string) *invoiceGrpc.VerifyChainResponse {
	return &invoiceGrpc.VerifyChainResponse{
		Success:       true,
		Code:          codes.Success,
		Ok:            false,
		BrokenDocId:   docID,
		BrokenDocType: docType,
		BrokenIndex:   index,
		Reason:        reason,
	}
}

func (s *Server) recomputeContentHash(ctx context.Context, userID, docType, docID string) (string, int32, error) {
	var (
		number              string
		issuedAt            sql.NullTime
		totalHT, totalVAT   int64
		totalTTC            int64
		vatExempt           bool
		originInvoiceNumber string
	)

	switch docType {
	case "INVOICE":
		err := s.db.QueryRowContext(ctx,
			`SELECT invoice_number, issued_at, total_ht_cents, total_vat_cents, total_ttc_cents, vat_exempt
			 FROM invoices WHERE invoice_id=$1 AND user_id=$2`,
			docID, userID,
		).Scan(&number, &issuedAt, &totalHT, &totalVAT, &totalTTC, &vatExempt)
		if err != nil {
			return "", codes.InternalError, err
		}
	case "CREDIT_NOTE":
		err := s.db.QueryRowContext(ctx,
			`SELECT cn.credit_note_number, cn.issued_at, cn.total_ht_cents, cn.total_vat_cents,
			        cn.total_ttc_cents, cn.vat_exempt, COALESCE(i.invoice_number, '')
			 FROM credit_notes cn
			 JOIN invoices i ON i.invoice_id = cn.invoice_id
			 WHERE cn.credit_note_id=$1 AND cn.user_id=$2`,
			docID, userID,
		).Scan(&number, &issuedAt, &totalHT, &totalVAT, &totalTTC, &vatExempt, &originInvoiceNumber)
		if err != nil {
			return "", codes.InternalError, err
		}
	default:
		return "", codes.InternalError, fmt.Errorf("unknown doc type %q", docType)
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return "", codes.InternalError, err
	}
	defer tx.Rollback()
	lines, err := loadSealLinesForDoc(ctx, tx, docType, docID)
	if err != nil {
		return "", codes.InternalError, err
	}

	return computeContentHash(sealableDoc{
		userID:              userID,
		docType:             docType,
		number:              number,
		issuedAt:            issuedAt.Time,
		totalHT:             totalHT,
		totalVAT:            totalVAT,
		totalTTC:            totalTTC,
		vatExempt:           vatExempt,
		originInvoiceNumber: originInvoiceNumber,
		lines:               lines,
	}), codes.Success, nil
}
