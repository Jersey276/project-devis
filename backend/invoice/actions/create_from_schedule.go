package actions

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"project-devis-invoice/actions/codes"
	invoiceGrpc "project-devis-invoice/services/grpc"
	scheduleGrpc "project-devis-invoice/services/schedulegrpc"
)

func (s *Server) CreateInvoiceFromSchedule(ctx context.Context, req *invoiceGrpc.CreateInvoiceFromScheduleRequest) (resp *invoiceGrpc.CreateInvoiceResponse, err error) {
	startedAt := time.Now()
	defer deferObserve("create_invoice_from_schedule", startedAt, func() (int32, bool) {
		if resp == nil {
			return codes.InternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	if req == nil {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var fieldErrors []*invoiceGrpc.ValidationError
	if strings.TrimSpace(req.UserId) == "" {
		fieldErrors = append(fieldErrors, &invoiceGrpc.ValidationError{Field: "user_id", Message: "Champ requis."})
	}
	if strings.TrimSpace(req.ScheduleId) == "" {
		fieldErrors = append(fieldErrors, &invoiceGrpc.ValidationError{Field: "schedule_id", Message: "Champ requis."})
	}
	if len(req.MonthIndexes) == 0 {
		fieldErrors = append(fieldErrors, &invoiceGrpc.ValidationError{Field: "month_indexes", Message: "Sélectionnez au moins un mois."})
	}
	if len(fieldErrors) > 0 {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InvalidInput, ValidationErrors: fieldErrors}, nil
	}

	// Load the schedule to confirm eligibility and resolve its quote/duration.
	schedResp, err := s.scheduleClient.GetSchedule(ctx, &scheduleGrpc.GetScheduleRequest{
		ScheduleId: req.ScheduleId,
		UserId:     req.UserId,
	})
	if err != nil {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InternalError}, err
	}
	if !schedResp.GetSuccess() || schedResp.GetSchedule() == nil {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.NotFound}, nil
	}
	sched := schedResp.GetSchedule()
	if !scheduleEligible(sched.GetStatus()) {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.SourceNotEligible}, nil
	}

	alreadyBilled, err := s.billedMonthsForSchedule(ctx, req.UserId, req.ScheduleId)
	if err != nil {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InternalError}, err
	}
	if code := validateMonthSelection(req.MonthIndexes, sched.GetDurationMonths(), alreadyBilled); code != codes.Success {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: code}, nil
	}

	invoiceID := uuid.New().String()
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO invoices (invoice_id, user_id, quote_id, schedule_id, billed_month_indexes, status)
		 VALUES ($1, $2, $3, $4, $5, 'DRAFT')`,
		invoiceID, req.UserId, sched.GetQuoteId(), req.ScheduleId, pq.Array(req.MonthIndexes),
	)
	if err != nil {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InternalError}, err
	}

	if req.IssueNow {
		return s.issue(ctx, invoiceID, req.UserId, req.SaleDate, req.DueInDays)
	}
	return &invoiceGrpc.CreateInvoiceResponse{Success: true, Code: codes.Success, InvoiceId: invoiceID}, nil
}

// billedMonthsForSchedule returns the union of month indexes already covered by
// ISSUED/PAID invoices of this schedule (CANCELLED frees the months again).
func (s *Server) billedMonthsForSchedule(ctx context.Context, userID, scheduleID string) ([]int32, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT billed_month_indexes FROM invoices
		 WHERE user_id=$1 AND schedule_id=$2 AND status IN ('ISSUED','PAID')`,
		userID, scheduleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []int32
	for rows.Next() {
		var months pq.Int32Array
		if err := rows.Scan(&months); err != nil {
			return nil, err
		}
		all = append(all, months...)
	}
	return all, rows.Err()
}
