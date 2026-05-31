package actions

import (
	"context"
	"database/sql"
	"strings"
	"time"

	scheduleGrpc "project-devis-schedule/services/grpc"
)

func (s *Server) GetSchedule(ctx context.Context, req *scheduleGrpc.GetScheduleRequest) (*scheduleGrpc.GetScheduleResponse, error) {
	startedAt := time.Now()
	var resp *scheduleGrpc.GetScheduleResponse
	var err error
	defer func() {
		code := CodeInternalError
		success := false
		if resp != nil {
			code = resp.Code
			success = resp.Success
		}
		recordOperation("get_schedule", success, code, startedAt, err)
	}()

	if req == nil || strings.TrimSpace(req.ScheduleId) == "" || strings.TrimSpace(req.UserId) == "" {
		resp = &scheduleGrpc.GetScheduleResponse{Success: false, Code: CodeInvalidInput}
		return resp, nil
	}

	var quoteID, status, name string
	var startMonth time.Time
	var durationMonths int32
	err = s.db.QueryRowContext(ctx,
		`SELECT quote_id, status, name, start_month, duration_months FROM schedules WHERE schedule_id=$1 AND user_id=$2`,
		req.ScheduleId, req.UserId,
	).Scan(&quoteID, &status, &name, &startMonth, &durationMonths)
	if err != nil {
		if err == sql.ErrNoRows {
			resp = &scheduleGrpc.GetScheduleResponse{Success: false, Code: CodeNotFound}
			return resp, nil
		}
		resp = &scheduleGrpc.GetScheduleResponse{Success: false, Code: CodeInternalError}
		return resp, err
	}

	lineRows, err := s.db.QueryContext(ctx, `
		SELECT sc.quote_line_id, COALESCE(SUM(sc.amount_cents), 0), COALESCE(ROUND(ql.unit_price * ql.quantity), 0)::BIGINT
		FROM schedule_cells sc
		JOIN quote_lines ql ON ql.line_id = sc.quote_line_id
		WHERE sc.schedule_id=$1
		GROUP BY sc.quote_line_id, ql.quantity, ql.unit_price
		ORDER BY sc.quote_line_id
	`, req.ScheduleId)
	if err != nil {
		resp = &scheduleGrpc.GetScheduleResponse{Success: false, Code: CodeInternalError}
		return resp, err
	}
	defer lineRows.Close()

	lineSummaries := make([]*scheduleGrpc.ScheduleLineSummary, 0)
	plannedTotalCents := int64(0)
	for lineRows.Next() {
		var lineID string
		var plannedCents, lineCents int64
		if err := lineRows.Scan(&lineID, &plannedCents, &lineCents); err != nil {
			resp = &scheduleGrpc.GetScheduleResponse{Success: false, Code: CodeInternalError}
			return resp, err
		}
		plannedTotalCents += plannedCents
		lineSummaries = append(lineSummaries, &scheduleGrpc.ScheduleLineSummary{
			QuoteLineId:   lineID,
			PlannedCents:  plannedCents,
			ExpectedCents: lineCents,
		})
	}
	if err := lineRows.Err(); err != nil {
		resp = &scheduleGrpc.GetScheduleResponse{Success: false, Code: CodeInternalError}
		return resp, err
	}

	columnRows, err := s.db.QueryContext(ctx,
		`SELECT month_index, COALESCE(SUM(amount_cents), 0) FROM schedule_cells WHERE schedule_id=$1 GROUP BY month_index ORDER BY month_index`,
		req.ScheduleId,
	)
	if err != nil {
		resp = &scheduleGrpc.GetScheduleResponse{Success: false, Code: CodeInternalError}
		return resp, err
	}
	defer columnRows.Close()

	columnTotals := make([]*scheduleGrpc.ScheduleColumnTotal, 0)
	var quoteTotalCents int64
	err = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(ROUND(unit_price * quantity)), 0)::BIGINT FROM quote_lines WHERE quote_id=$1`,
		quoteID,
	).Scan(&quoteTotalCents)
	if err != nil {
		resp = &scheduleGrpc.GetScheduleResponse{Success: false, Code: CodeInternalError}
		return resp, err
	}

	for columnRows.Next() {
		var monthIndex int32
		var amountCents int64
		if err := columnRows.Scan(&monthIndex, &amountCents); err != nil {
			resp = &scheduleGrpc.GetScheduleResponse{Success: false, Code: CodeInternalError}
			return resp, err
		}
		columnTotals = append(columnTotals, &scheduleGrpc.ScheduleColumnTotal{
			MonthIndex:  monthIndex,
			AmountCents: amountCents,
		})
	}
	if err := columnRows.Err(); err != nil {
		resp = &scheduleGrpc.GetScheduleResponse{Success: false, Code: CodeInternalError}
		return resp, err
	}

	resp = &scheduleGrpc.GetScheduleResponse{
		Success: true,
		Code:    CodeSuccess,
		Schedule: &scheduleGrpc.ScheduleDetails{
			ScheduleId:        req.ScheduleId,
			QuoteId:           quoteID,
			Status:            status,
			Name:              name,
			StartMonth:        startMonth.Format("2006-01"),
			DurationMonths:    durationMonths,
			Lines:             lineSummaries,
			ColumnTotals:      columnTotals,
			QuoteTotalCents:   quoteTotalCents,
			PlannedTotalCents: plannedTotalCents,
		},
	}

	return resp, nil
}
