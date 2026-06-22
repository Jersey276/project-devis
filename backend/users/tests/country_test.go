package tests

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"project-devis-users/actions"
	usersGrpc "project-devis-users/services/grpc"
)

func TestCreateCountry_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`INSERT INTO countries`).
		WithArgs("FR", "France").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	resp, err := srv.CreateCountry(context.Background(), &usersGrpc.CreateCountryRequest{Code: "FR", Name: "France"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.CountryId != 1 {
		t.Fatalf("expected country_id 1, got %d", resp.CountryId)
	}
}

func TestCreateCountry_AlreadyExists(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`INSERT INTO countries`).
		WithArgs("FR", "France").
		WillReturnError(&pq.Error{Code: "23505"})

	resp, err := srv.CreateCountry(context.Background(), &usersGrpc.CreateCountryRequest{Code: "FR", Name: "France"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for duplicate code")
	}
	if resp.Code != actions.CodeAlreadyExists {
		t.Fatalf("expected CodeAlreadyExists, got %d", resp.Code)
	}
}

func TestCreateCountry_InvalidCode(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateCountry(context.Background(), &usersGrpc.CreateCountryRequest{Code: "FRA", Name: "France"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for 3-char code")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestListCountries_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT id, code, name, is_eu FROM countries`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "is_eu"}).
			AddRow(1, "FR", "France", true).
			AddRow(2, "DE", "Germany", true))

	resp, err := srv.ListCountries(context.Background(), &usersGrpc.ListCountriesRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if len(resp.Countries) != 2 {
		t.Fatalf("expected 2 countries, got %d", len(resp.Countries))
	}
}

func TestDeleteCountry_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`DELETE FROM countries`).
		WithArgs(int32(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, err := srv.DeleteCountry(context.Background(), &usersGrpc.DeleteCountryRequest{CountryId: 999})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for non-existent country")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
}
