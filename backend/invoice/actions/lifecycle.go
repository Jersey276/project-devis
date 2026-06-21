package actions

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"project-devis-invoice/actions/codes"
	invoiceGrpc "project-devis-invoice/services/grpc"
)

// lifecycleTransitions is the source of truth for the e-invoicing status flow
// (réforme FR B2B). Strict forward path NONE→DEPOSITED→RECEIVED→APPROVED→COLLECTED;
// REJECTED is reachable from any active non-terminal state. No backward moves, no
// skips, no self-loops. REJECTED and COLLECTED are terminal.
var lifecycleTransitions = map[string]map[string]bool{
	"NONE":      {"DEPOSITED": true},
	"DEPOSITED": {"RECEIVED": true, "REJECTED": true},
	"RECEIVED":  {"APPROVED": true, "REJECTED": true},
	"APPROVED":  {"COLLECTED": true, "REJECTED": true},
	"REJECTED":  {},
	"COLLECTED": {},
}

// nextLifecycleAllowed reports whether moving from current to target is a legal
// manual transition. target must be one of the five real statuses (not NONE).
func nextLifecycleAllowed(current, target string) bool {
	return lifecycleTransitions[current][target]
}

func (s *Server) SetInvoiceLifecycleStatus(ctx context.Context, req *invoiceGrpc.SetInvoiceLifecycleStatusRequest) (resp *invoiceGrpc.GenericResponse, err error) {
	startedAt := time.Now()
	defer deferObserve("set_invoice_lifecycle_status", startedAt, func() (int32, bool) {
		if resp == nil {
			return codes.InternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	if req == nil || strings.TrimSpace(req.InvoiceId) == "" || strings.TrimSpace(req.UserId) == "" {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	target := strings.TrimSpace(req.GetStatus())
	// target must be a real status; NONE is the sentinel, never a transition target.
	if _, ok := lifecycleTransitions[target]; !ok || target == "NONE" {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	defer tx.Rollback()

	if code, err := applyLifecycleTransition(ctx, tx, req.InvoiceId, req.UserId, target, strings.TrimSpace(req.GetNote())); code != codes.Success {
		return &invoiceGrpc.GenericResponse{Success: false, Code: code}, err
	}

	if err := tx.Commit(); err != nil {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	return &invoiceGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}

// applyLifecycleTransition is the single source of truth for advancing an
// invoice's e-invoicing lifecycle inside a caller-owned tx: lock the row, guard
// status∈ISSUED|PAID, validate the move against lifecycleTransitions, then update
// the status and append the event. The caller commits. Returns codes.Success on
// success, otherwise the business code (and any DB error). Both the manual RPC
// and the PDP deposit flow go through here so guards and the append-only log
// cannot be bypassed.
func applyLifecycleTransition(ctx context.Context, tx *sql.Tx, invoiceID, userID, target, note string) (int32, error) {
	var status, lifecycle string
	err := tx.QueryRowContext(ctx,
		`SELECT status, lifecycle_status FROM invoices
		 WHERE invoice_id=$1 AND user_id=$2 FOR UPDATE`,
		invoiceID, userID,
	).Scan(&status, &lifecycle)
	if err == sql.ErrNoRows {
		return codes.NotFound, nil
	}
	if err != nil {
		return codes.InternalError, err
	}

	if status != "ISSUED" && status != "PAID" {
		return codes.LifecycleRequiresIssued, nil
	}
	if !nextLifecycleAllowed(lifecycle, target) {
		return codes.LifecycleTransitionInvalid, nil
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE invoices SET lifecycle_status=$1, updated_at=NOW() WHERE invoice_id=$2 AND user_id=$3`,
		target, invoiceID, userID,
	); err != nil {
		return codes.InternalError, err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO invoice_lifecycle_events (invoice_id, user_id, status, note)
		 VALUES ($1, $2, $3, $4)`,
		invoiceID, userID, target, note,
	); err != nil {
		return codes.InternalError, err
	}
	return codes.Success, nil
}

func (s *Server) ListInvoiceLifecycleEvents(ctx context.Context, req *invoiceGrpc.ListInvoiceLifecycleEventsRequest) (resp *invoiceGrpc.ListInvoiceLifecycleEventsResponse, err error) {
	startedAt := time.Now()
	defer deferObserve("list_invoice_lifecycle_events", startedAt, func() (int32, bool) {
		if resp == nil {
			return codes.InternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	if req == nil || strings.TrimSpace(req.InvoiceId) == "" || strings.TrimSpace(req.UserId) == "" {
		return &invoiceGrpc.ListInvoiceLifecycleEventsResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	// Ownership filtered in the same query: a non-owned invoice yields an empty
	// history, same as a real empty one. The detail page already 404s via GetInvoice.
	rows, err := s.db.QueryContext(ctx,
		`SELECT e.status, e.note, e.created_at FROM invoice_lifecycle_events e
		 WHERE e.invoice_id=$1
		   AND EXISTS (SELECT 1 FROM invoices i
		               WHERE i.invoice_id=e.invoice_id AND i.user_id=$2)
		 ORDER BY e.created_at, e.id`,
		req.InvoiceId, req.UserId,
	)
	if err != nil {
		return &invoiceGrpc.ListInvoiceLifecycleEventsResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	events := make([]*invoiceGrpc.InvoiceLifecycleEvent, 0)
	for rows.Next() {
		var status, note string
		var createdAt time.Time
		if err := rows.Scan(&status, &note, &createdAt); err != nil {
			return &invoiceGrpc.ListInvoiceLifecycleEventsResponse{Success: false, Code: codes.InternalError}, err
		}
		events = append(events, &invoiceGrpc.InvoiceLifecycleEvent{
			Status:    status,
			Note:      note,
			CreatedAt: createdAt.Format(time.RFC3339),
		})
	}
	if err := rows.Err(); err != nil {
		return &invoiceGrpc.ListInvoiceLifecycleEventsResponse{Success: false, Code: codes.InternalError}, err
	}
	return &invoiceGrpc.ListInvoiceLifecycleEventsResponse{Success: true, Code: codes.Success, Events: events}, nil
}
