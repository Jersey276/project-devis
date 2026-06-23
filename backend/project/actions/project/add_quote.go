package project

import (
	"context"
	"database/sql"

	"project-devis-project/actions/codes"
	projectGrpc "project-devis-project/services/grpc"
)

func AddQuote(ctx context.Context, db *sql.DB, req *projectGrpc.AddQuoteToProjectRequest) (*projectGrpc.GenericResponse, error) {
	if req.ProjectId == "" || req.UserId == "" || req.QuoteId == "" {
		return &projectGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	// Verify project ownership
	var exists bool
	if err := db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM projects WHERE project_id = $1 AND user_id = $2)`,
		req.ProjectId, req.UserId,
	).Scan(&exists); err != nil {
		return &projectGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if !exists {
		return &projectGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}

	res, err := db.ExecContext(ctx,
		`INSERT INTO project_quotes (project_id, quote_id) VALUES ($1, $2) ON CONFLICT (quote_id) DO NOTHING`,
		req.ProjectId, req.QuoteId,
	)
	if err != nil {
		return &projectGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &projectGrpc.GenericResponse{Success: false, Code: codes.AlreadyExists}, nil
	}

	return &projectGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
