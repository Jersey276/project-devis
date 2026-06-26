package tests

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	projectGrpc "project-devis-project/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

var projectCols = []string{
	"project_id", "user_id", "name", "client_id", "status",
	"created_at", "updated_at", "quote_count",
}

func addProjectRow(rows *sqlmock.Rows, id, name string) *sqlmock.Rows {
	return rows.AddRow(id, "user-1", name, sql.NullString{Valid: false}, "active",
		"2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z", int32(0))
}

func TestListProjects_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\)`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))

	rows := addProjectRow(
		addProjectRow(sqlmock.NewRows(projectCols), "proj-1", "Projet A"),
		"proj-2", "Projet B",
	)
	mock.ExpectQuery(`SELECT p.project_id`).
		WillReturnRows(rows)

	resp, err := srv.ListProjects(context.Background(), &projectGrpc.ListProjectsRequest{
		UserId:   "user-1",
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Total != 2 {
		t.Fatalf("expected total 2, got %d", resp.Total)
	}
	if len(resp.Projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(resp.Projects))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListProjects_MissingUserID(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.ListProjects(context.Background(), &projectGrpc.ListProjectsRequest{UserId: ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure")
	}
	if resp.Code != 1003 {
		t.Fatalf("expected InvalidInput=1003, got %d", resp.Code)
	}
}

func TestListProjects_WithSearch(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\)`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery(`SELECT p.project_id`).
		WillReturnRows(addProjectRow(sqlmock.NewRows(projectCols), "proj-1", "Projet Matching"))

	resp, err := srv.ListProjects(context.Background(), &projectGrpc.ListProjectsRequest{
		UserId:   "user-1",
		Search:   "Matching",
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Total != 1 {
		t.Fatalf("expected total 1, got %d", resp.Total)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListProjects_SortByName(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\)`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mock.ExpectQuery(`SELECT p.project_id`).
		WillReturnRows(sqlmock.NewRows(projectCols))

	resp, err := srv.ListProjects(context.Background(), &projectGrpc.ListProjectsRequest{
		UserId:        "user-1",
		SortBy:        "name",
		SortDirection: "ASC",
		Page:          1,
		PageSize:      20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

func TestListProjects_SortByUnknown(t *testing.T) {
	srv, mock := setupServer(t)

	// sort_by="unknown" doit fallback sur p.created_at — la query doit quand même s'exécuter
	mock.ExpectQuery(`SELECT COUNT\(\*\)`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mock.ExpectQuery(`SELECT p.project_id`).
		WillReturnRows(sqlmock.NewRows(projectCols))

	resp, err := srv.ListProjects(context.Background(), &projectGrpc.ListProjectsRequest{
		UserId:   "user-1",
		SortBy:   "unknown_field",
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

func TestListProjects_DBError(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\)`).
		WillReturnError(fmt.Errorf("db timeout"))

	resp, err := srv.ListProjects(context.Background(), &projectGrpc.ListProjectsRequest{
		UserId:   "user-1",
		Page:     1,
		PageSize: 20,
	})
	if err == nil {
		t.Fatal("expected error to be propagated")
	}
	if resp.Success {
		t.Fatal("expected failure")
	}
	if resp.Code != 2001 {
		t.Fatalf("expected InternalError=2001, got %d", resp.Code)
	}
}
