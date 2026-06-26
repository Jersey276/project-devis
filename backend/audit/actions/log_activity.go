package actions

import (
	"context"

	auditGrpc "project-devis-audit/services/grpc"
)

func (s *Server) LogActivity(ctx context.Context, req *auditGrpc.LogActivityRequest) (*auditGrpc.LogActivityResponse, error) {
	userID := nullableString(req.UserId)
	reqBody := nullableString(req.ReqBody)

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO activity_logs (user_id, method, url, duration_ms, req_body, resp_body, resp_status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID, req.Method, req.Url, req.DurationMs, reqBody, req.RespBody, req.RespStatus,
	)
	if err != nil {
		return &auditGrpc.LogActivityResponse{Success: false, Code: CodeInternalError}, nil
	}
	return &auditGrpc.LogActivityResponse{Success: true, Code: CodeSuccess}, nil
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
