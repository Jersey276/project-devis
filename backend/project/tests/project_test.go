package tests

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	projectGrpc "project-devis-project/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

// CreateProject

func TestCreateProject_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`INSERT INTO projects`).
		WithArgs(sqlmock.AnyArg(), "user-1", "Mon Projet", nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.CreateProject(context.Background(), &projectGrpc.CreateProjectRequest{
		UserId: "user-1",
		Name:   "Mon Projet",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.ProjectId == "" {
		t.Fatal("expected a project_id")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateProject_MissingFields(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.CreateProject(context.Background(), &projectGrpc.CreateProjectRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure")
	}
	if resp.Code != 1003 {
		t.Fatalf("expected InvalidInput=1003, got %d", resp.Code)
	}
	if len(resp.ValidationErrors) == 0 {
		t.Fatal("expected validation errors")
	}
}

func TestCreateProject_WithClientID(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`INSERT INTO projects`).
		WithArgs(sqlmock.AnyArg(), "user-1", "Projet Client", "client-abc").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.CreateProject(context.Background(), &projectGrpc.CreateProjectRequest{
		UserId:   "user-1",
		Name:     "Projet Client",
		ClientId: "client-abc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// GetProject

func TestGetProject_Success(t *testing.T) {
	srv, mock := setupServer(t)

	rows := sqlmock.NewRows([]string{
		"project_id", "user_id", "name", "client_id", "status",
		"created_at", "updated_at", "quote_count",
	}).AddRow("proj-1", "user-1", "Mon Projet", sql.NullString{Valid: false}, "active",
		"2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z", int32(0))

	mock.ExpectQuery(`SELECT p.project_id`).
		WithArgs("proj-1", "user-1").
		WillReturnRows(rows)

	resp, err := srv.GetProject(context.Background(), &projectGrpc.GetProjectRequest{
		ProjectId: "proj-1",
		UserId:    "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Project.Name != "Mon Projet" {
		t.Fatalf("expected name 'Mon Projet', got %q", resp.Project.Name)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetProject_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT p.project_id`).
		WithArgs("proj-x", "user-1").
		WillReturnError(sql.ErrNoRows)

	resp, err := srv.GetProject(context.Background(), &projectGrpc.GetProjectRequest{
		ProjectId: "proj-x",
		UserId:    "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure")
	}
	if resp.Code != 1001 {
		t.Fatalf("expected NotFound=1001, got %d", resp.Code)
	}
}

func TestGetProject_MissingInput(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.GetProject(context.Background(), &projectGrpc.GetProjectRequest{UserId: "user-1"})
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

// UpdateProject

func TestUpdateProject_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE projects`).
		WithArgs("Nouveau Nom", nil, "active", "proj-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.UpdateProject(context.Background(), &projectGrpc.UpdateProjectRequest{
		ProjectId: "proj-1",
		UserId:    "user-1",
		Name:      "Nouveau Nom",
		Status:    "active",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateProject_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE projects`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, err := srv.UpdateProject(context.Background(), &projectGrpc.UpdateProjectRequest{
		ProjectId: "proj-x",
		UserId:    "user-1",
		Name:      "Nom",
		Status:    "active",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure")
	}
	if resp.Code != 1001 {
		t.Fatalf("expected NotFound=1001, got %d", resp.Code)
	}
}

func TestUpdateProject_MissingName(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.UpdateProject(context.Background(), &projectGrpc.UpdateProjectRequest{
		ProjectId: "proj-1",
		UserId:    "user-1",
		Name:      "",
	})
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

// DeleteProject

func TestDeleteProject_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("proj-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM project_quotes`).
		WithArgs("proj-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM projects`).
		WithArgs("proj-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	resp, err := srv.DeleteProject(context.Background(), &projectGrpc.DeleteProjectRequest{
		ProjectId: "proj-1",
		UserId:    "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteProject_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("proj-x", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	resp, err := srv.DeleteProject(context.Background(), &projectGrpc.DeleteProjectRequest{
		ProjectId: "proj-x",
		UserId:    "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure")
	}
	if resp.Code != 1001 {
		t.Fatalf("expected NotFound=1001, got %d", resp.Code)
	}
}

func TestDeleteProject_MissingInput(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.DeleteProject(context.Background(), &projectGrpc.DeleteProjectRequest{
		ProjectId: "",
		UserId:    "user-1",
	})
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

func TestDeleteProject_DBError(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("proj-1", "user-1").
		WillReturnError(fmt.Errorf("db down"))

	resp, err := srv.DeleteProject(context.Background(), &projectGrpc.DeleteProjectRequest{
		ProjectId: "proj-1",
		UserId:    "user-1",
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
