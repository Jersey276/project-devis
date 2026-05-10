package tests

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"project-devis-users/actions"
	usersGrpc "project-devis-users/services/grpc"
)

func TestCreateUser_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), "new@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.CreateUser(context.Background(), &usersGrpc.CreateUserRequest{Email: "new@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.UserId == "" {
		t.Fatal("expected non-empty user_id")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateUser_MissingEmail(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateUser(context.Background(), &usersGrpc.CreateUserRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for missing email")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestCreateUser_AlreadyExists(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), "existing@example.com").
		WillReturnError(&pq.Error{Code: "23505"})

	resp, err := srv.CreateUser(context.Background(), &usersGrpc.CreateUserRequest{Email: "existing@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for existing email")
	}
	if resp.Code != actions.CodeAlreadyExists {
		t.Fatalf("expected CodeAlreadyExists, got %d", resp.Code)
	}
}

func TestGetUser_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT user_id, email`).
		WithArgs("user-123").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "email", "phone", "company", "siren", "vat"}).
			AddRow("user-123", "user@example.com", "", "", "", ""))

	resp, err := srv.GetUser(context.Background(), &usersGrpc.GetUserRequest{UserId: "user-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.User.UserId != "user-123" {
		t.Fatalf("expected user_id user-123, got %q", resp.User.UserId)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT user_id, email`).
		WithArgs("unknown").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "email", "phone", "company", "siren", "vat"}))

	resp, err := srv.GetUser(context.Background(), &usersGrpc.GetUserRequest{UserId: "unknown"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for unknown user")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
}

func TestUpdateUser_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE users SET phone`).
		WithArgs("0600000000", "ACME", "123456789", "FR12345", "user-123").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.UpdateUser(context.Background(), &usersGrpc.UpdateUserRequest{
		UserId:  "user-123",
		Phone:   "0600000000",
		Company: "ACME",
		Siren:   "123456789",
		Vat:     "FR12345",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

func TestDeleteUser_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM addresses WHERE owner_type='client'`).
		WithArgs("user-123").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`DELETE FROM addresses WHERE owner_type='user'`).
		WithArgs("user-123").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`DELETE FROM users`).
		WithArgs("user-123").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	resp, err := srv.DeleteUser(context.Background(), &usersGrpc.DeleteUserRequest{UserId: "user-123"})
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

// TestDeleteUser_DeletesClientAddresses is a regression test for the
// cascade-ordering bug: client-owned addresses must be deleted before the
// user row, otherwise the ON DELETE CASCADE on clients.user_id wipes the
// client rows first and the subquery resolves to nothing.
func TestDeleteUser_DeletesClientAddresses(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM addresses WHERE owner_type='client'`).
		WithArgs("user-123").
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(`DELETE FROM addresses WHERE owner_type='user'`).
		WithArgs("user-123").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM users`).
		WithArgs("user-123").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	resp, err := srv.DeleteUser(context.Background(), &usersGrpc.DeleteUserRequest{UserId: "user-123"})
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

func TestDeleteUser_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM addresses WHERE owner_type='client'`).
		WithArgs("ghost").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`DELETE FROM addresses WHERE owner_type='user'`).
		WithArgs("ghost").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`DELETE FROM users`).
		WithArgs("ghost").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	resp, err := srv.DeleteUser(context.Background(), &usersGrpc.DeleteUserRequest{UserId: "ghost"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for non-existent user")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
