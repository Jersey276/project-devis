package project

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"project-devis-project/actions/codes"
	projectGrpc "project-devis-project/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *projectGrpc.CreateProjectRequest) (*projectGrpc.CreateProjectResponse, error) {
	var fieldErrors []*projectGrpc.ValidationError

	if req.UserId == "" {
		fieldErrors = append(fieldErrors, &projectGrpc.ValidationError{Field: "user_id", Message: "Champ requis."})
	}
	if req.Name == "" {
		fieldErrors = append(fieldErrors, &projectGrpc.ValidationError{Field: "name", Message: "Champ requis."})
	}

	if len(fieldErrors) > 0 {
		return &projectGrpc.CreateProjectResponse{Success: false, Code: codes.InvalidInput, ValidationErrors: fieldErrors}, nil
	}

	projectID := uuid.New().String()
	_, err := db.ExecContext(ctx,
		`INSERT INTO projects (project_id, user_id, name, client_id, status) VALUES ($1, $2, $3, $4, 'active')`,
		projectID, req.UserId, req.Name, nullableString(req.ClientId),
	)
	if err != nil {
		return &projectGrpc.CreateProjectResponse{Success: false, Code: codes.InternalError}, err
	}

	return &projectGrpc.CreateProjectResponse{Success: true, Code: codes.Success, ProjectId: projectID}, nil
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
