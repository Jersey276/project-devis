package tests

import (
	"context"
	"testing"

	projectGrpc "project-devis-project/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

// AddQuoteToProject

func TestAddQuoteToProject_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("proj-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO project_quotes`).
		WithArgs("proj-1", "quote-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.AddQuoteToProject(context.Background(), &projectGrpc.AddQuoteToProjectRequest{
		ProjectId: "proj-1",
		UserId:    "user-1",
		QuoteId:   "quote-1",
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

func TestAddQuoteToProject_AlreadyExists(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("proj-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO project_quotes`).
		WithArgs("proj-1", "quote-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, err := srv.AddQuoteToProject(context.Background(), &projectGrpc.AddQuoteToProjectRequest{
		ProjectId: "proj-1",
		UserId:    "user-1",
		QuoteId:   "quote-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure")
	}
	if resp.Code != 1002 {
		t.Fatalf("expected AlreadyExists=1002, got %d", resp.Code)
	}
}

func TestAddQuoteToProject_ProjectNotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("proj-x", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	resp, err := srv.AddQuoteToProject(context.Background(), &projectGrpc.AddQuoteToProjectRequest{
		ProjectId: "proj-x",
		UserId:    "user-1",
		QuoteId:   "quote-1",
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

func TestAddQuoteToProject_MissingInput(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.AddQuoteToProject(context.Background(), &projectGrpc.AddQuoteToProjectRequest{
		ProjectId: "proj-1",
		UserId:    "user-1",
		QuoteId:   "",
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

// RemoveQuoteFromProject

func TestRemoveQuoteFromProject_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("proj-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`DELETE FROM project_quotes`).
		WithArgs("proj-1", "quote-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.RemoveQuoteFromProject(context.Background(), &projectGrpc.RemoveQuoteFromProjectRequest{
		ProjectId: "proj-1",
		UserId:    "user-1",
		QuoteId:   "quote-1",
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

func TestRemoveQuoteFromProject_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("proj-x", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	resp, err := srv.RemoveQuoteFromProject(context.Background(), &projectGrpc.RemoveQuoteFromProjectRequest{
		ProjectId: "proj-x",
		UserId:    "user-1",
		QuoteId:   "quote-1",
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

// ListProjectQuoteIds

func TestListProjectQuoteIds_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT pq.quote_id`).
		WithArgs("proj-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id"}).
			AddRow("q-1").
			AddRow("q-2"))

	resp, err := srv.ListProjectQuoteIds(context.Background(), &projectGrpc.ListProjectQuoteIdsRequest{
		ProjectId: "proj-1",
		UserId:    "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if len(resp.QuoteIds) != 2 {
		t.Fatalf("expected 2 quote ids, got %d", len(resp.QuoteIds))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListProjectQuoteIds_Empty(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT pq.quote_id`).
		WithArgs("proj-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id"}))

	resp, err := srv.ListProjectQuoteIds(context.Background(), &projectGrpc.ListProjectQuoteIdsRequest{
		ProjectId: "proj-1",
		UserId:    "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if len(resp.QuoteIds) != 0 {
		t.Fatalf("expected empty slice, got %v", resp.QuoteIds)
	}
}

func TestListProjectQuoteIds_MissingInput(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.ListProjectQuoteIds(context.Background(), &projectGrpc.ListProjectQuoteIdsRequest{
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
