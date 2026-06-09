package actions

import (
	"context"
	"log"

	emailGrpc "project-devis-email/services/grpc"
)

func (s *Server) GetEmailLogs(ctx context.Context, req *emailGrpc.GetEmailLogsRequest) (*emailGrpc.GetEmailLogsResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	var total int32
	countErr := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM email_logs WHERE ($1::text IS NULL OR user_id = $1)",
		nullableString(req.UserId),
	).Scan(&total)
	if countErr != nil {
		log.Printf("count email logs failed: %v", countErr)
		return &emailGrpc.GetEmailLogsResponse{}, nil
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, to_email, type, COALESCE(reference_name, ''), status,
		        opened_at IS NOT NULL, clicked_at IS NOT NULL,
		        to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		 FROM email_logs
		 WHERE ($1::text IS NULL OR user_id = $1)
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		nullableString(req.UserId), limit, offset,
	)
	if err != nil {
		return &emailGrpc.GetEmailLogsResponse{}, err
	}
	defer rows.Close()

	var logs []*emailGrpc.EmailLog
	for rows.Next() {
		var entry emailGrpc.EmailLog
		var id int
		if err := rows.Scan(
			&id, &entry.ToEmail, &entry.Type, &entry.ReferenceName,
			&entry.Status, &entry.Opened, &entry.Clicked, &entry.CreatedAt,
		); err != nil {
			log.Printf("scan email log row failed: %v", err)
			continue
		}
		entry.Id = int32(id)
		logs = append(logs, &entry)
	}

	if err := rows.Err(); err != nil {
		return &emailGrpc.GetEmailLogsResponse{}, err
	}

	return &emailGrpc.GetEmailLogsResponse{Logs: logs, Total: total}, nil
}

