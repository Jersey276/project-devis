package project

import (
	"context"
	"database/sql"

	"project-devis-project/actions/codes"
	projectGrpc "project-devis-project/services/grpc"
)

func ListQuoteIds(ctx context.Context, db *sql.DB, req *projectGrpc.ListProjectQuoteIdsRequest) (*projectGrpc.ListProjectQuoteIdsResponse, error) {
	if req.ProjectId == "" || req.UserId == "" {
		return &projectGrpc.ListProjectQuoteIdsResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	rows, err := db.QueryContext(ctx,
		`SELECT pq.quote_id FROM project_quotes pq
		 JOIN projects p ON p.project_id = pq.project_id
		 WHERE pq.project_id = $1 AND p.user_id = $2`,
		req.ProjectId, req.UserId,
	)
	if err != nil {
		return &projectGrpc.ListProjectQuoteIdsResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	var quoteIds []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return &projectGrpc.ListProjectQuoteIdsResponse{Success: false, Code: codes.InternalError}, err
		}
		quoteIds = append(quoteIds, id)
	}
	if err := rows.Err(); err != nil {
		return &projectGrpc.ListProjectQuoteIdsResponse{Success: false, Code: codes.InternalError}, err
	}

	return &projectGrpc.ListProjectQuoteIdsResponse{Success: true, Code: codes.Success, QuoteIds: quoteIds}, nil
}
