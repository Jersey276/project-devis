package project

import (
	"context"
	"database/sql"

	"project-devis-project/actions/codes"
	projectGrpc "project-devis-project/services/grpc"
)

func Delete(ctx context.Context, db *sql.DB, req *projectGrpc.DeleteProjectRequest) (*projectGrpc.GenericResponse, error) {
	if req.ProjectId == "" || req.UserId == "" {
		return &projectGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	// Verify ownership first
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM projects WHERE project_id = $1 AND user_id = $2)`,
		req.ProjectId, req.UserId,
	).Scan(&exists)
	if err != nil {
		return &projectGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if !exists {
		return &projectGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return &projectGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM project_quotes WHERE project_id = $1`, req.ProjectId); err != nil {
		return &projectGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM projects WHERE project_id = $1 AND user_id = $2`, req.ProjectId, req.UserId); err != nil {
		return &projectGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}

	if err := tx.Commit(); err != nil {
		return &projectGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}

	return &projectGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
