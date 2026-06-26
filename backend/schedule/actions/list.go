package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	scheduleGrpc "project-devis-schedule/services/grpc"
)

func (s *Server) ListSchedules(ctx context.Context, req *scheduleGrpc.ListSchedulesRequest) (*scheduleGrpc.ListSchedulesResponse, error) {
	startedAt := time.Now()
	var resp *scheduleGrpc.ListSchedulesResponse
	var err error
	defer deferObserve("list_schedules", startedAt, func() (int32, bool) {
		if resp == nil {
			return CodeInternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	clientID := ""
	if req.GetFilters() != nil {
		clientID = strings.TrimSpace(req.GetFilters().ClientId)
	}
	if req == nil || (strings.TrimSpace(req.UserId) == "" && clientID == "") {
		resp = &scheduleGrpc.ListSchedulesResponse{Success: false, Code: CodeInvalidInput}
		return resp, nil
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	where, args := buildScheduleFilters(req.UserId, req.QuoteId, req.QuoteIds, req.Filters)

	var total int64
	if err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schedules"+where, args...).Scan(&total); err != nil {
		resp = &scheduleGrpc.ListSchedulesResponse{Success: false, Code: CodeInternalError}
		return resp, err
	}

	orderBy := buildScheduleOrderBy(req.SortBy, req.SortDirection)

	args = append(args, pageSize, offset)
	n := len(args)
	query := fmt.Sprintf(
		`SELECT schedule_id, quote_id, status, name, start_month, duration_months, COALESCE(client_id, '')
		 FROM schedules%s ORDER BY %s LIMIT $%d OFFSET $%d`,
		where, orderBy, n-1, n,
	)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		resp = &scheduleGrpc.ListSchedulesResponse{Success: false, Code: CodeInternalError}
		return resp, err
	}
	defer rows.Close()

	schedules := make([]*scheduleGrpc.ScheduleSummary, 0)
	for rows.Next() {
		var scheduleID, quoteID, status, name, clientID string
		var startMonth time.Time
		var durationMonths int32
		if err := rows.Scan(&scheduleID, &quoteID, &status, &name, &startMonth, &durationMonths, &clientID); err != nil {
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
			ClientId:       clientID,
		})
	}
	if err = rows.Err(); err != nil {
		resp = &scheduleGrpc.ListSchedulesResponse{Success: false, Code: CodeInternalError}
		return resp, err
	}

	resp = &scheduleGrpc.ListSchedulesResponse{
		Success:   true,
		Code:      CodeSuccess,
		Schedules: schedules,
		Total:     total,
	}
	return resp, nil
}

var allowedScheduleSortColumns = map[string]string{
	"id":             "schedule_id",
	"name":           "name",
	"quoteId":        "quote_id",
	"status":         "status",
	"startMonth":     "start_month",
	"durationMonths": "duration_months",
}

func buildScheduleOrderBy(sortBy, sortDirection string) string {
	col, ok := allowedScheduleSortColumns[sortBy]
	if !ok {
		col = "created_at"
	}
	if strings.ToUpper(sortDirection) == "ASC" {
		return col + " ASC"
	}
	return col + " DESC"
}

func buildScheduleFilters(userID, quoteID string, quoteIDs []string, f *scheduleGrpc.ScheduleFilters) (string, []interface{}) {
	// client_id filter is only active when user_id is empty (customer-mode calls
	// from the gateway). A non-empty user_id always takes precedence so that a
	// provider cannot bypass ownership by sending an arbitrary client_id.
	clientID := ""
	if f != nil && strings.TrimSpace(userID) == "" {
		clientID = strings.TrimSpace(f.ClientId)
	}

	var args []interface{}
	var clauses []string

	if clientID != "" {
		args = append(args, clientID)
		clauses = append(clauses, "client_id = $1")
	} else {
		args = append(args, userID)
		clauses = append(clauses, "user_id = $1")
	}

	if strings.TrimSpace(quoteID) != "" {
		args = append(args, quoteID)
		clauses = append(clauses, fmt.Sprintf("quote_id = $%d", len(args)))
	}

	if len(quoteIDs) > 0 {
		placeholders := make([]string, len(quoteIDs))
		for i, id := range quoteIDs {
			args = append(args, id)
			placeholders[i] = fmt.Sprintf("$%d", len(args))
		}
		clauses = append(clauses, "quote_id IN ("+strings.Join(placeholders, ",")+")")
	}

	if f != nil {
		if len(f.Statuses) > 0 {
			placeholders := make([]string, len(f.Statuses))
			for i, st := range f.Statuses {
				args = append(args, st)
				placeholders[i] = fmt.Sprintf("$%d", len(args))
			}
			clauses = append(clauses, "status IN ("+strings.Join(placeholders, ",")+")")
		}
		if f.StartFrom != "" {
			args = append(args, f.StartFrom)
			clauses = append(clauses, fmt.Sprintf("start_month >= $%d", len(args)))
		}
		if f.StartTo != "" {
			args = append(args, f.StartTo)
			clauses = append(clauses, fmt.Sprintf("start_month <= $%d", len(args)))
		}
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}
