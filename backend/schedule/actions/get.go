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

	quoteID, status, name, startMonth, durationMonths, err := loadScheduleHeader(ctx, s.db, req.ScheduleId, req.UserId)
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

	lineSummaries, plannedTotalCents, err := loadScheduleLines(ctx, s.db, req.ScheduleId, expectedByLineID)
	if err != nil {
		resp = &scheduleGrpc.GetScheduleResponse{Success: false, Code: CodeInternalError}
		return resp, err
	}

	columnTotals, err := loadScheduleColumns(ctx, s.db, req.ScheduleId)
	if err != nil {
		resp = &scheduleGrpc.GetScheduleResponse{Success: false, Code: CodeInternalError}
		return resp, err
	}

	var quoteTotalCents int64
	for _, expected := range expectedByLineID {
		quoteTotalCents += expected
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

func loadScheduleHeader(ctx context.Context, db *sql.DB, scheduleID, userID string) (quoteID, status, name string, startMonth time.Time, durationMonths int32, err error) {
	err = db.QueryRowContext(ctx,
		`SELECT quote_id, status, name, start_month, duration_months FROM schedules WHERE schedule_id=$1 AND user_id=$2`,
		scheduleID, userID,
	).Scan(&quoteID, &status, &name, &startMonth, &durationMonths)
	return
}

func loadScheduleLines(ctx context.Context, db *sql.DB, scheduleID string, expectedByLineID map[string]int64) ([]*scheduleGrpc.ScheduleLineSummary, int64, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT sc.quote_line_id, COALESCE(SUM(sc.amount_cents), 0)
		FROM schedule_cells sc
		WHERE sc.schedule_id=$1
		GROUP BY sc.quote_line_id
		ORDER BY sc.quote_line_id
	`, scheduleID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	summaries := make([]*scheduleGrpc.ScheduleLineSummary, 0)
	var plannedTotal int64
	for rows.Next() {
		var lineID string
		var plannedCents int64
		if err := rows.Scan(&lineID, &plannedCents); err != nil {
			return nil, 0, err
		}
		plannedTotal += plannedCents
		summaries = append(summaries, &scheduleGrpc.ScheduleLineSummary{
			QuoteLineId:   lineID,
			PlannedCents:  plannedCents,
			ExpectedCents: expectedByLineID[lineID],
		})
	}
	return summaries, plannedTotal, rows.Err()
}

func loadScheduleColumns(ctx context.Context, db *sql.DB, scheduleID string) ([]*scheduleGrpc.ScheduleColumnTotal, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT month_index, COALESCE(SUM(amount_cents), 0) FROM schedule_cells WHERE schedule_id=$1 GROUP BY month_index ORDER BY month_index`,
		scheduleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totals := make([]*scheduleGrpc.ScheduleColumnTotal, 0)
	for rows.Next() {
		var monthIndex int32
		var amountCents int64
		if err := rows.Scan(&monthIndex, &amountCents); err != nil {
			return nil, err
		}
		totals = append(totals, &scheduleGrpc.ScheduleColumnTotal{
			MonthIndex:  monthIndex,
			AmountCents: amountCents,
		})
	}
	return totals, rows.Err()
}

// GetScheduleCells returns the raw per-(line, month) amounts of a schedule.
// Downstream services (invoicing) need this granularity to bill a subset of
// months with a correct per-line breakdown — GetSchedule only exposes totals.
func (s *Server) GetScheduleCells(ctx context.Context, req *scheduleGrpc.GetScheduleCellsRequest) (*scheduleGrpc.GetScheduleCellsResponse, error) {
	startedAt := time.Now()
	var resp *scheduleGrpc.GetScheduleCellsResponse
	var err error
	defer deferObserve("get_schedule_cells", startedAt, func() (int32, bool) {
		if resp == nil {
			return CodeInternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	if req == nil || strings.TrimSpace(req.ScheduleId) == "" || strings.TrimSpace(req.UserId) == "" {
		resp = &scheduleGrpc.GetScheduleCellsResponse{Success: false, Code: CodeInvalidInput}
		return resp, nil
	}

	// Ownership guard: the schedule must belong to the caller.
	var exists bool
	err = s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM schedules WHERE schedule_id=$1 AND user_id=$2)`,
		req.ScheduleId, req.UserId,
	).Scan(&exists)
	if err != nil {
		resp = &scheduleGrpc.GetScheduleCellsResponse{Success: false, Code: CodeInternalError}
		return resp, err
	}
	if !exists {
		resp = &scheduleGrpc.GetScheduleCellsResponse{Success: false, Code: CodeNotFound}
		return resp, nil
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT quote_line_id, month_index, amount_cents
		 FROM schedule_cells WHERE schedule_id=$1
		 ORDER BY quote_line_id, month_index`,
		req.ScheduleId,
	)
	if err != nil {
		resp = &scheduleGrpc.GetScheduleCellsResponse{Success: false, Code: CodeInternalError}
		return resp, err
	}
	defer rows.Close()

	cells := make([]*scheduleGrpc.ScheduleCell, 0)
	for rows.Next() {
		var lineID string
		var monthIndex int32
		var amountCents int64
		if err := rows.Scan(&lineID, &monthIndex, &amountCents); err != nil {
			resp = &scheduleGrpc.GetScheduleCellsResponse{Success: false, Code: CodeInternalError}
			return resp, err
		}
		cells = append(cells, &scheduleGrpc.ScheduleCell{
			QuoteLineId: lineID,
			MonthIndex:  monthIndex,
			AmountCents: amountCents,
		})
	}
	if err := rows.Err(); err != nil {
		resp = &scheduleGrpc.GetScheduleCellsResponse{Success: false, Code: CodeInternalError}
		return resp, err
	}

	resp = &scheduleGrpc.GetScheduleCellsResponse{Success: true, Code: CodeSuccess, Cells: cells}
	return resp, nil
}
