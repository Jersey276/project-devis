package project

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"project-devis-project/actions/codes"
	projectGrpc "project-devis-project/services/grpc"
)

func List(ctx context.Context, db *sql.DB, req *projectGrpc.ListProjectsRequest) (*projectGrpc.ListProjectsResponse, error) {
	if req.UserId == "" {
		return &projectGrpc.ListProjectsResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	where, args := buildProjectFilters(req)

	var total int64
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM projects p"+where, args...).Scan(&total); err != nil {
		return &projectGrpc.ListProjectsResponse{Success: false, Code: codes.InternalError}, err
	}

	orderBy := buildProjectOrderBy(req.SortBy, req.SortDirection)

	args = append(args, pageSize, offset)
	n := len(args)
	query := fmt.Sprintf(
		`SELECT p.project_id, p.user_id, p.name, p.client_id, p.status,
		        TO_CHAR(p.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
		        TO_CHAR(p.updated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
		        COUNT(pq.quote_id) AS quote_count
		 FROM projects p
		 LEFT JOIN project_quotes pq ON pq.project_id = p.project_id
		 %s
		 GROUP BY p.project_id, p.user_id, p.name, p.client_id, p.status, p.created_at, p.updated_at
		 ORDER BY %s LIMIT $%d OFFSET $%d`,
		where, orderBy, n-1, n,
	)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return &projectGrpc.ListProjectsResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	var projects []*projectGrpc.Project
	for rows.Next() {
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
		if err := rows.Scan(&projectID, &userID, &name, &clientID, &status, &createdAt, &updatedAt, &quoteCount); err != nil {
			return &projectGrpc.ListProjectsResponse{Success: false, Code: codes.InternalError}, err
		}
		projects = append(projects, &projectGrpc.Project{
			ProjectId:  projectID,
			UserId:     userID,
			Name:       name,
			ClientId:   clientID.String,
			Status:     status,
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
			QuoteCount: quoteCount,
		})
	}
	if err := rows.Err(); err != nil {
		return &projectGrpc.ListProjectsResponse{Success: false, Code: codes.InternalError}, err
	}

	return &projectGrpc.ListProjectsResponse{Success: true, Code: codes.Success, Projects: projects, Total: total}, nil
}

var allowedSortColumns = map[string]string{
	"name":       "p.name",
	"status":     "p.status",
	"created_at": "p.created_at",
	"updated_at": "p.updated_at",
}

func buildProjectOrderBy(sortBy, sortDirection string) string {
	col, ok := allowedSortColumns[sortBy]
	if !ok {
		col = "p.created_at"
	}
	if strings.ToUpper(sortDirection) == "ASC" {
		return col + " ASC"
	}
	return col + " DESC"
}

func buildProjectFilters(req *projectGrpc.ListProjectsRequest) (string, []interface{}) {
	args := []interface{}{req.UserId}
	clauses := []string{"p.user_id = $1"}

	if req.Search != "" {
		args = append(args, "%"+req.Search+"%")
		clauses = append(clauses, fmt.Sprintf("p.name ILIKE $%d", len(args)))
	}
	if req.Status != "" {
		args = append(args, req.Status)
		clauses = append(clauses, fmt.Sprintf("p.status = $%d", len(args)))
	}
	if req.ClientId != "" {
		args = append(args, req.ClientId)
		clauses = append(clauses, fmt.Sprintf("p.client_id = $%d", len(args)))
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}
