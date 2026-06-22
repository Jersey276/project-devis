package actions

import (
	"context"
	"fmt"
	"strings"

	auditGrpc "project-devis-audit/services/grpc"
)

func (s *Server) ListActivityLogs(ctx context.Context, req *auditGrpc.ListActivityLogsRequest) (*auditGrpc.ListActivityLogsResponse, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	where, args := buildFilters(req.Filters)

	countQuery := "SELECT COUNT(*) FROM activity_logs" + where
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return &auditGrpc.ListActivityLogsResponse{Success: false, Code: CodeInternalError}, nil
	}

	args = append(args, pageSize, offset)
	n := len(args)
	dataQuery := `SELECT id, COALESCE(user_id,''), method, url, duration_ms,
	                     resp_status,
	                     to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
	              FROM activity_logs` + where +
		fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, n-1, n)

	rows, err := s.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return &auditGrpc.ListActivityLogsResponse{Success: false, Code: CodeInternalError}, nil
	}
	defer rows.Close()

	var logs []*auditGrpc.ActivityLog
	for rows.Next() {
		var l auditGrpc.ActivityLog
		if err := rows.Scan(&l.Id, &l.UserId, &l.Method, &l.Url, &l.DurationMs,
			&l.RespStatus, &l.CreatedAt); err != nil {
			return &auditGrpc.ListActivityLogsResponse{Success: false, Code: CodeInternalError}, nil
		}
		logs = append(logs, &l)
	}

	return &auditGrpc.ListActivityLogsResponse{
		Success: true,
		Code:    CodeSuccess,
		Logs:    logs,
		Total:   total,
	}, nil
}

func buildFilters(f *auditGrpc.ActivityLogFilters) (string, []interface{}) {
	if f == nil {
		return "", nil
	}
	var clauses []string
	var args []interface{}
	n := 1

	if f.UserId != "" {
		args = append(args, f.UserId)
		clauses = append(clauses, fmt.Sprintf("user_id = $%d", n))
		n++
	}
	if f.UrlContains != "" {
		args = append(args, "%"+f.UrlContains+"%")
		clauses = append(clauses, fmt.Sprintf("url ILIKE $%d", n))
		n++
	}
	if f.RespStatus != 0 {
		args = append(args, f.RespStatus)
		clauses = append(clauses, fmt.Sprintf("resp_status = $%d", n))
		n++
	}
	if f.DateFrom != "" {
		args = append(args, f.DateFrom)
		clauses = append(clauses, fmt.Sprintf("created_at >= $%d", n))
		n++
	}
	if f.DateTo != "" {
		args = append(args, f.DateTo)
		clauses = append(clauses, fmt.Sprintf("created_at <= $%d", n))
		n++
	}

	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}
