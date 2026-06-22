package actions

import (
	"context"
	"database/sql"

	auditGrpc "project-devis-audit/services/grpc"
)

func (s *Server) GetActivityLog(ctx context.Context, req *auditGrpc.GetActivityLogRequest) (*auditGrpc.GetActivityLogResponse, error) {
	var l auditGrpc.ActivityLogDetail
	err := s.db.QueryRowContext(ctx,
		`SELECT id, COALESCE(user_id,''), method, url, duration_ms,
		        COALESCE(req_body,''), resp_body, resp_status,
		        to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		 FROM activity_logs WHERE id = $1`,
		req.Id,
	).Scan(&l.Id, &l.UserId, &l.Method, &l.Url, &l.DurationMs,
		&l.ReqBody, &l.RespBody, &l.RespStatus, &l.CreatedAt)

	if err == sql.ErrNoRows {
		return &auditGrpc.GetActivityLogResponse{Success: false, Code: CodeNotFound}, nil
	}
	if err != nil {
		return &auditGrpc.GetActivityLogResponse{Success: false, Code: CodeInternalError}, nil
	}

	return &auditGrpc.GetActivityLogResponse{Success: true, Code: CodeSuccess, Log: &l}, nil
}
