package actions

import (
	"context"

	auditGrpc "project-devis-audit/services/grpc"
)

func (s *Server) GetActivityStats(ctx context.Context, _ *auditGrpc.GetActivityStatsRequest) (*auditGrpc.GetActivityStatsResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS day,
		        resp_status,
		        COUNT(*) AS count
		 FROM activity_logs
		 WHERE created_at >= now() - INTERVAL '6 months'
		 GROUP BY day, resp_status
		 ORDER BY day, resp_status`,
	)
	if err != nil {
		return &auditGrpc.GetActivityStatsResponse{Success: false, Code: CodeInternalError}, nil
	}
	defer rows.Close()

	var stats []*auditGrpc.StatusCount
	for rows.Next() {
		var sc auditGrpc.StatusCount
		if err := rows.Scan(&sc.Date, &sc.RespStatus, &sc.Count); err != nil {
			return &auditGrpc.GetActivityStatsResponse{Success: false, Code: CodeInternalError}, nil
		}
		stats = append(stats, &sc)
	}

	return &auditGrpc.GetActivityStatsResponse{
		Success: true,
		Code:    CodeSuccess,
		Stats:   stats,
	}, nil
}
