package project

import (
	"context"
	"database/sql"

	"project-devis-project/actions/codes"
	projectGrpc "project-devis-project/services/grpc"
)

func Get(ctx context.Context, db *sql.DB, req *projectGrpc.GetProjectRequest) (*projectGrpc.GetProjectResponse, error) {
	if req.ProjectId == "" || req.UserId == "" {
		return &projectGrpc.GetProjectResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var (
		projectID  string
		userID     string
		name       string
		clientID   sql.NullString
		status     string
		createdAt  string
		updatedAt  string
		quoteCount int32
	)

	err := db.QueryRowContext(ctx,
		`SELECT p.project_id, p.user_id, p.name, p.client_id, p.status,
		        TO_CHAR(p.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
		        TO_CHAR(p.updated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
		        COUNT(pq.quote_id)
		 FROM projects p
		 LEFT JOIN project_quotes pq ON pq.project_id = p.project_id
		 WHERE p.project_id = $1 AND p.user_id = $2
		 GROUP BY p.project_id, p.user_id, p.name, p.client_id, p.status, p.created_at, p.updated_at`,
		req.ProjectId, req.UserId,
	).Scan(&projectID, &userID, &name, &clientID, &status, &createdAt, &updatedAt, &quoteCount)

	if err == sql.ErrNoRows {
		return &projectGrpc.GetProjectResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &projectGrpc.GetProjectResponse{Success: false, Code: codes.InternalError}, err
	}

	return &projectGrpc.GetProjectResponse{
		Success: true,
		Code:    codes.Success,
		Project: &projectGrpc.Project{
			ProjectId:  projectID,
			UserId:     userID,
			Name:       name,
			ClientId:   clientID.String,
			Status:     status,
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
			QuoteCount: quoteCount,
		},
	}, nil
}
