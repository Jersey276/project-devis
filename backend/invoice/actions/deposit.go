package actions

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"project-devis-invoice/actions/codes"
	"project-devis-invoice/pdp"
	invoiceGrpc "project-devis-invoice/services/grpc"
)

// DepositInvoice deposits an issued invoice on the e-invoicing platform (B6) and
// advances its lifecycle to DEPOSITED. The platform call happens first; only on
// success do we move the lifecycle, through applyLifecycleTransition so the
// ISSUED|PAID guard, the transition table and the append-only event log all
// apply (a double deposit is rejected as DEPOSITED→DEPOSITED). The platform
// handle, if any, is persisted in the same tx.
func (s *Server) DepositInvoice(ctx context.Context, req *invoiceGrpc.DepositInvoiceRequest) (resp *invoiceGrpc.GenericResponse, err error) {
	startedAt := time.Now()
	defer deferObserve("deposit_invoice", startedAt, func() (int32, bool) {
		if resp == nil {
			return codes.InternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	if req == nil || strings.TrimSpace(req.InvoiceId) == "" || strings.TrimSpace(req.UserId) == "" {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	// Pre-check ownership and issued state before calling the platform: fail fast,
	// and never submit a draft. applyLifecycleTransition re-checks under FOR UPDATE.
	var status string
	var number sql.NullString // NULL while DRAFT
	err = s.db.QueryRowContext(ctx,
		`SELECT status, invoice_number FROM invoices WHERE invoice_id=$1 AND user_id=$2`,
		req.InvoiceId, req.UserId,
	).Scan(&status, &number)
	if err == sql.ErrNoRows {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if status != "ISSUED" && status != "PAID" {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.LifecycleRequiresIssued}, nil
	}

	// Resolve the recipient against the e-invoicing directory before depositing:
	// we never route to a recipient the directory cannot place. The recipient
	// SIRET is the frozen client_siret of the legal snapshot (B4). The no-op
	// directory resolves everyone with an empty handle, so this is inert until a
	// real annuaire is wired.
	var recipientSiret sql.NullString
	err = s.db.QueryRowContext(ctx,
		`SELECT client_siret FROM invoice_party_snapshots WHERE invoice_id=$1`,
		req.InvoiceId,
	).Scan(&recipientSiret)
	if err != nil && err != sql.ErrNoRows { // no snapshot: resolve with an empty SIRET
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	routing, err := s.pdpDirectory.Resolve(ctx, recipientSiret.String)
	if err != nil {
		if errors.Is(err, pdp.ErrRecipientNotFound) {
			return &invoiceGrpc.GenericResponse{Success: false, Code: codes.RecipientNotInDirectory}, nil
		}
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}

	result, err := s.pdpClient.Submit(ctx, pdp.SubmitInput{
		InvoiceID:     req.InvoiceId,
		UserID:        req.UserId,
		InvoiceNumber: number.String,
	})
	if err != nil {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.PDPSubmissionFailed}, nil
	}
	target, ok := pdp.ToLifecycleStatus(result.Status)
	if !ok || target != "DEPOSITED" {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.PDPSubmissionFailed}, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	defer tx.Rollback()

	if code, err := applyLifecycleTransition(ctx, tx, req.InvoiceId, req.UserId, target, strings.TrimSpace(req.GetNote())); code != codes.Success {
		return &invoiceGrpc.GenericResponse{Success: false, Code: code}, err
	}
	// Freeze both platform handles in the same tx: the submission id (if the
	// platform assigned one) and the directory routing id (if the annuaire
	// returned one). Either may be empty under the no-op adapters; we only persist
	// non-empty values so frozen columns stay NULL rather than "".
	sid := strings.TrimSpace(result.SubmissionID)
	routingID := strings.TrimSpace(routing.RoutingID)
	if sid != "" || routingID != "" {
		if _, err := tx.ExecContext(ctx,
			`UPDATE invoices
			    SET pdp_submission_id    = COALESCE(NULLIF($1, ''), pdp_submission_id),
			        recipient_routing_id = COALESCE(NULLIF($2, ''), recipient_routing_id)
			  WHERE invoice_id=$3 AND user_id=$4`,
			sid, routingID, req.InvoiceId, req.UserId,
		); err != nil {
			return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	return &invoiceGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
