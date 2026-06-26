package project

import (
	"context"
	"database/sql"

	"project-devis-project/actions/codes"
	projectGrpc "project-devis-project/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *projectGrpc.UpdateProjectRequest) (*projectGrpc.GenericResponse, error) {
	if req.ProjectId == "" || req.UserId == "" {
		return &projectGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	if req.Name == "" {
		return &projectGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	result, err := db.ExecContext(ctx,
		`UPDATE projects SET name = $1, client_id = $2, status = $3, updated_at = NOW()
		 WHERE project_id = $4 AND user_id = $5`,
		req.Name, nullableString(req.ClientId), req.Status, req.ProjectId, req.UserId,
	)
	if err != nil {
		return &projectGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return &projectGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &projectGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
