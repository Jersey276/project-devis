package actions

import (
	"context"
	"strings"
	"time"

	scheduleGrpc "project-devis-schedule/services/grpc"
)

func (s *Server) ListSchedules(ctx context.Context, req *scheduleGrpc.ListSchedulesRequest) (*scheduleGrpc.ListSchedulesResponse, error) {
	startedAt := time.Now()
	var resp *scheduleGrpc.ListSchedulesResponse
	var err error
	defer func() {
		code := CodeInternalError
		success := false
		if resp != nil {
			code = resp.Code
			success = resp.Success
		}
		recordOperation("list_schedules", success, code, startedAt, err)
	}()

	if req == nil || strings.TrimSpace(req.UserId) == "" {
		resp = &scheduleGrpc.ListSchedulesResponse{Success: false, Code: CodeInvalidInput}
		return resp, nil
	}

	baseQuery := `SELECT schedule_id, quote_id, status, name, start_month, duration_months FROM schedules WHERE user_id=$1`
	args := []any{req.UserId}
	if strings.TrimSpace(req.QuoteId) != "" {
		baseQuery += ` AND quote_id=$2`
		args = append(args, req.QuoteId)
	}
	baseQuery += ` ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		resp = &scheduleGrpc.ListSchedulesResponse{Success: false, Code: CodeInternalError}
		return resp, err
	}
	defer rows.Close()

	schedules := make([]*scheduleGrpc.ScheduleSummary, 0)
	for rows.Next() {
		var scheduleID, quoteID, status, name string
		var startMonth time.Time
		var durationMonths int32
		if err := rows.Scan(&scheduleID, &quoteID, &status, &name, &startMonth, &durationMonths); err != nil {
			resp = &scheduleGrpc.ListSchedulesResponse{Success: false, Code: CodeInternalError}
			return resp, err
		}
		schedules = append(schedules, &scheduleGrpc.ScheduleSummary{
			ScheduleId:     scheduleID,
			QuoteId:        quoteID,
			Status:         status,
			Name:           name,
			StartMonth:     startMonth.Format("2006-01"),
			DurationMonths: durationMonths,
		})
	}
	if err := rows.Err(); err != nil {
		resp = &scheduleGrpc.ListSchedulesResponse{Success: false, Code: CodeInternalError}
		return resp, err
	}

	resp = &scheduleGrpc.ListSchedulesResponse{
		Success:   true,
		Code:      CodeSuccess,
		Schedules: schedules,
	}

	return resp, nil
}
