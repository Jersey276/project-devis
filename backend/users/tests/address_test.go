package tests

import (
	"context"
	"testing"

	"project-devis-users/actions"
	usersGrpc "project-devis-users/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateAddress_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`INSERT INTO addresses`).
		WithArgs("user-1", "Home", "1 rue de la Paix", nil, "Paris", "75001", int32(1), nil, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))

	resp, err := srv.CreateAddress(context.Background(), &usersGrpc.CreateAddressRequest{
		UserId:    "user-1",
		Name:      "Home",
		Street:    "1 rue de la Paix",
		City:      "Paris",
		ZipCode:   "75001",
		CountryId: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.AddressId != 42 {
		t.Fatalf("expected address_id 42, got %d", resp.AddressId)
	}
}

func TestCreateAddress_MissingRequired(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateAddress(context.Background(), &usersGrpc.CreateAddressRequest{
		UserId: "user-1",
		// missing Name, Street, City, ZipCode, CountryId
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

func TestListAddresses_Success(t *testing.T) {
	srv, mock := setupServer(t)

	cols := []string{"id", "user_id", "name", "street", "additional_street", "city", "zip_code", "country_id", "email", "phone", "archived"}
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(1, "user-1", "Home", "1 rue de la Paix", nil, "Paris", "75001", 1, nil, nil, false).
			AddRow(2, "user-1", "Office", "5 avenue du Général", nil, "Lyon", "69001", 1, nil, nil, false))

	resp, err := srv.ListAddresses(context.Background(), &usersGrpc.ListAddressesRequest{UserId: "user-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if len(resp.Addresses) != 2 {
		t.Fatalf("expected 2 addresses, got %d", len(resp.Addresses))
	}
}

func TestArchiveAddress_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE addresses SET archived_at`).
		WithArgs(int32(1), "user-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.ArchiveAddress(context.Background(), &usersGrpc.ArchiveAddressRequest{
		AddressId: 1,
		UserId:    "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

func TestArchiveAddress_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE addresses SET archived_at`).
		WithArgs(int32(99), "user-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, err := srv.ArchiveAddress(context.Background(), &usersGrpc.ArchiveAddressRequest{
		AddressId: 99,
		UserId:    "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for not-found address")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
}
