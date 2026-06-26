package actions

import (
	"context"
	"fmt"
	"log"
	"strings"

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

	where, args := buildEmailLogsWhere(req)

	var total int32
	countSQL := "SELECT COUNT(*) FROM email_logs WHERE " + where
	if err := s.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		log.Printf("count email logs failed: %v", err)
		return &emailGrpc.GetEmailLogsResponse{}, nil
	}

	// Append LIMIT and OFFSET as the last two args.
	limitArg := len(args) + 1
	offsetArg := len(args) + 2
	querySQL := fmt.Sprintf(
		`SELECT id, to_email, type, COALESCE(reference_name, ''), status,
		        opened_at IS NOT NULL, clicked_at IS NOT NULL,
		        to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		 FROM email_logs
		 WHERE %s
		 ORDER BY created_at DESC
		 LIMIT $%d OFFSET $%d`,
		where, limitArg, offsetArg,
	)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, querySQL, args...)
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

// buildEmailLogsWhere builds the WHERE clause and args slice for email_logs queries.
// Arg $1 is always user_id; additional filters are appended as needed.
func buildEmailLogsWhere(req *emailGrpc.GetEmailLogsRequest) (string, []any) {
	args := []any{nullableString(req.UserId)}
	conditions := []string{"($1::text IS NULL OR user_id = $1)"}

	next := func() string {
		args = append(args, nil) // placeholder, replaced below
		return fmt.Sprintf("$%d", len(args))
	}

	statuses := filterEmpty(req.Statuses)
	if len(statuses) > 0 {
		ph := next()
		args[len(args)-1] = statuses
		conditions = append(conditions, fmt.Sprintf("status = ANY(%s::text[])", ph))
	}

	types := filterEmpty(req.Types)
	if len(types) > 0 {
		ph := next()
		args[len(args)-1] = types
		conditions = append(conditions, fmt.Sprintf("type = ANY(%s::text[])", ph))
	}

	if req.DateFrom != "" {
		ph := next()
		args[len(args)-1] = req.DateFrom
		conditions = append(conditions, fmt.Sprintf("created_at >= %s::date", ph))
	}

	if req.DateTo != "" {
		ph := next()
		args[len(args)-1] = req.DateTo
		conditions = append(conditions, fmt.Sprintf("created_at < %s::date + interval '1 day'", ph))
	}

	if req.Search != "" {
		ph := next()
		args[len(args)-1] = "%" + req.Search + "%"
		conditions = append(conditions, fmt.Sprintf("(to_email ILIKE %s OR reference_name ILIKE %s)", ph, ph))
	}

	return strings.Join(conditions, " AND "), args
}

// filterEmpty removes blank strings from a slice (avoids empty splits from the gateway).
func filterEmpty(ss []string) []string {
	out := ss[:0]
	for _, s := range ss {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
