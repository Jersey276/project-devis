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
	defer deferObserve("get_schedule", startedAt, func() (int32, bool) {
		if resp == nil {
			return CodeInternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

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

	expectedByLineID, err := getQuoteLineExpectedCents(ctx, req.UserId, quoteID)
	if err != nil {
		resp = &scheduleGrpc.GetScheduleResponse{Success: false, Code: CodeInternalError}
		return resp, err
	}

	lineRows, err := s.db.QueryContext(ctx, `
		SELECT sc.quote_line_id, COALESCE(SUM(sc.amount_cents), 0)
		FROM schedule_cells sc
		WHERE sc.schedule_id=$1
		GROUP BY sc.quote_line_id
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
		var plannedCents int64
		if err := lineRows.Scan(&lineID, &plannedCents); err != nil {
			resp = &scheduleGrpc.GetScheduleResponse{Success: false, Code: CodeInternalError}
			return resp, err
		}
		plannedTotalCents += plannedCents
		lineSummaries = append(lineSummaries, &scheduleGrpc.ScheduleLineSummary{
			QuoteLineId:   lineID,
			PlannedCents:  plannedCents,
			ExpectedCents: expectedByLineID[lineID],
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
	for _, expected := range expectedByLineID {
		quoteTotalCents += expected
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
