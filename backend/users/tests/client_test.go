package tests

import (
	"context"
	"testing"

	"project-devis-users/actions"
	usersGrpc "project-devis-users/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateClient_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`INSERT INTO clients`).
		WithArgs(sqlmock.AnyArg(), "user-1", "Jean", "Dupont", "jean@example.com", nil, "Acme", nil, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.CreateClient(context.Background(), &usersGrpc.CreateClientRequest{
		UserId:    "user-1",
		FirstName: "Jean",
		LastName:  "Dupont",
		Email:     "jean@example.com",
		Company:   "Acme",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.ClientId == "" {
		t.Fatal("expected non-empty client_id")
	}
}

func TestCreateClient_MissingRequired(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateClient(context.Background(), &usersGrpc.CreateClientRequest{
		UserId: "user-1",
		// missing FirstName, LastName
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for missing fields")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestCreateClient_MissingUserID(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateClient(context.Background(), &usersGrpc.CreateClientRequest{
		FirstName: "Jean",
		LastName:  "Dupont",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for missing user_id")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestGetClient_Success(t *testing.T) {
	srv, mock := setupServer(t)

	cols := []string{"client_id", "user_id", "first_name", "last_name", "email", "phone", "company", "siren", "vat", "archived"}
	mock.ExpectQuery(`SELECT client_id, user_id`).
		WithArgs("c-1", "user-1").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("c-1", "user-1", "Jean", "Dupont", "jean@example.com", nil, "Acme", nil, nil, false))

	resp, err := srv.GetClient(context.Background(), &usersGrpc.GetClientRequest{
		ClientId: "c-1",
		UserId:   "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Client == nil || resp.Client.ClientId != "c-1" {
		t.Fatalf("expected client c-1, got %+v", resp.Client)
	}
	if resp.Client.Email != "jean@example.com" || resp.Client.Company != "Acme" {
		t.Fatalf("unexpected client fields: %+v", resp.Client)
	}
}

func TestGetClient_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	cols := []string{"client_id", "user_id", "first_name", "last_name", "email", "phone", "company", "siren", "vat", "archived"}
	mock.ExpectQuery(`SELECT client_id, user_id`).
		WithArgs("ghost", "user-1").
		WillReturnRows(sqlmock.NewRows(cols))

	resp, err := srv.GetClient(context.Background(), &usersGrpc.GetClientRequest{
		ClientId: "ghost",
		UserId:   "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for not-found client")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
}

// TestGetClient_ExcludesArchived asserts that the WHERE clause filters
// archived rows. Without this, the gateway's owner-resolver would treat
// archived clients as valid address owners.
func TestGetClient_ExcludesArchived(t *testing.T) {
	srv, mock := setupServer(t)

	cols := []string{"client_id", "user_id", "first_name", "last_name", "email", "phone", "company", "siren", "vat", "archived"}
	mock.ExpectQuery(`SELECT client_id, user_id.*archived_at IS NULL`).
		WithArgs("c-1", "user-1").
		WillReturnRows(sqlmock.NewRows(cols))

	resp, err := srv.GetClient(context.Background(), &usersGrpc.GetClientRequest{
		ClientId: "c-1",
		UserId:   "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for archived client")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListClients_ExcludesArchivedByDefault(t *testing.T) {
	srv, mock := setupServer(t)

	cols := []string{"client_id", "user_id", "first_name", "last_name", "email", "phone", "company", "siren", "vat", "archived"}
	mock.ExpectQuery(`SELECT client_id, user_id.*FROM clients WHERE user_id=\$1 AND archived_at IS NULL`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("c-1", "user-1", "Jean", "Dupont", nil, nil, nil, nil, nil, false).
			AddRow("c-2", "user-1", "Marie", "Martin", nil, nil, nil, nil, nil, false))

	resp, err := srv.ListClients(context.Background(), &usersGrpc.ListClientsRequest{UserId: "user-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if len(resp.Clients) != 2 {
		t.Fatalf("expected 2 clients, got %d", len(resp.Clients))
	}
}

func TestListClients_IncludeArchived(t *testing.T) {
	srv, mock := setupServer(t)

	cols := []string{"client_id", "user_id", "first_name", "last_name", "email", "phone", "company", "siren", "vat", "archived"}
	// IncludeArchived=true must NOT add the archived_at IS NULL filter.
	mock.ExpectQuery(`SELECT client_id, user_id.*FROM clients WHERE user_id=\$1 ORDER`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("c-1", "user-1", "Jean", "Dupont", nil, nil, nil, nil, nil, false).
			AddRow("c-2", "user-1", "Marie", "Martin", nil, nil, nil, nil, nil, true))

	resp, err := srv.ListClients(context.Background(), &usersGrpc.ListClientsRequest{
		UserId:          "user-1",
		IncludeArchived: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if len(resp.Clients) != 2 {
		t.Fatalf("expected 2 clients, got %d", len(resp.Clients))
	}
}

func TestUpdateClient_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE clients SET first_name`).
		WithArgs("Jean", "Dupont", "jean@example.com", nil, nil, nil, nil, "c-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.UpdateClient(context.Background(), &usersGrpc.UpdateClientRequest{
		ClientId:  "c-1",
		UserId:    "user-1",
		FirstName: "Jean",
		LastName:  "Dupont",
		Email:     "jean@example.com",
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

func TestUpdateClient_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE clients SET first_name`).
		WithArgs("Jean", "Dupont", nil, nil, nil, nil, nil, "ghost", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, err := srv.UpdateClient(context.Background(), &usersGrpc.UpdateClientRequest{
		ClientId:  "ghost",
		UserId:    "user-1",
		FirstName: "Jean",
		LastName:  "Dupont",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for not-found client")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
}

func TestUpdateClient_MissingRequired(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.UpdateClient(context.Background(), &usersGrpc.UpdateClientRequest{
		ClientId: "c-1",
		UserId:   "user-1",
		// missing FirstName, LastName
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for missing fields")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestArchiveClient_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE clients SET archived_at`).
		WithArgs("c-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.ArchiveClient(context.Background(), &usersGrpc.ArchiveClientRequest{
		ClientId: "c-1",
		UserId:   "user-1",
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

func TestArchiveClient_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE clients SET archived_at`).
		WithArgs("ghost", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, err := srv.ArchiveClient(context.Background(), &usersGrpc.ArchiveClientRequest{
		ClientId: "ghost",
		UserId:   "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for not-found client")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
}
