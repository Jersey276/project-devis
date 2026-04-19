package tests

import (
	"context"
	"testing"

	"project-devis-users/actions"
	usersGrpc "project-devis-users/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateCountryGroup_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`INSERT INTO country_groups`).
		WithArgs("Zone EU").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	resp, err := srv.CreateCountryGroup(context.Background(), &usersGrpc.CreateCountryGroupRequest{Name: "Zone EU"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.CountryGroupId != 1 {
		t.Fatalf("expected country_group_id 1, got %d", resp.CountryGroupId)
	}
}

func TestCreateCountryGroup_MissingName(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateCountryGroup(context.Background(), &usersGrpc.CreateCountryGroupRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for missing name")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestAttachCountry_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`INSERT INTO country_group_countries`).
		WithArgs(int32(1), int32(2)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.AttachCountry(context.Background(), &usersGrpc.AttachCountryRequest{
		CountryGroupId: 1,
		CountryId:      2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

func TestDetachCountry_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`DELETE FROM country_group_countries`).
		WithArgs(int32(1), int32(2)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.DetachCountry(context.Background(), &usersGrpc.DetachCountryRequest{
		CountryGroupId: 1,
		CountryId:      2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

func TestDetachCountry_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`DELETE FROM country_group_countries`).
		WithArgs(int32(1), int32(99)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, err := srv.DetachCountry(context.Background(), &usersGrpc.DetachCountryRequest{
		CountryGroupId: 1,
		CountryId:      99,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for non-existent relation")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
}

func TestDeleteCountryGroup_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`DELETE FROM country_groups`).
		WithArgs(int32(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.DeleteCountryGroup(context.Background(), &usersGrpc.DeleteCountryGroupRequest{CountryGroupId: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}
