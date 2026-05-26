package tests

import (
	"context"
	"testing"

	"project-devis-users/actions"
	usersGrpc "project-devis-users/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

const authUser = "user-1"

func TestCreateAddress_Success(t *testing.T) {
	srv, mock := setupServer(t)

	// INSERT-SELECT with the auth predicate baked in. Args: owner_type,
	// owner_id, name, street, additional_street, city, zip_code, country_id,
	// email, phone, auth_user_id.
	mock.ExpectQuery(`INSERT INTO addresses`).
		WithArgs("user", authUser, "Home", "1 rue de la Paix", nil, "Paris", "75001", int32(1), nil, nil, authUser).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))

	resp, err := srv.CreateAddress(context.Background(), &usersGrpc.CreateAddressRequest{
		OwnerType:  usersGrpc.OwnerType_OWNER_TYPE_USER,
		OwnerId:    authUser,
		Name:       "Home",
		Street:     "1 rue de la Paix",
		City:       "Paris",
		ZipCode:    "75001",
		CountryId:  1,
		AuthUserId: authUser,
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
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateAddress_MissingRequired(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateAddress(context.Background(), &usersGrpc.CreateAddressRequest{
		OwnerType:  usersGrpc.OwnerType_OWNER_TYPE_USER,
		OwnerId:    authUser,
		AuthUserId: authUser,
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

func TestCreateAddress_InvalidOwnerType(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.CreateAddress(context.Background(), &usersGrpc.CreateAddressRequest{
		OwnerType:  usersGrpc.OwnerType_OWNER_TYPE_UNSPECIFIED,
		OwnerId:    authUser,
		Name:       "Home",
		Street:     "1 rue de la Paix",
		City:       "Paris",
		ZipCode:    "75001",
		CountryId:  1,
		AuthUserId: authUser,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for invalid owner_type")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
}

// TestCreateAddress_RejectsForeignOwner: the INSERT-SELECT returns no rows
// when the (owner_type, owner_id, auth) triple doesn't resolve. The action
// must surface NotFound, not InternalError.
func TestCreateAddress_RejectsForeignOwner(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`INSERT INTO addresses`).
		WithArgs("client", "c-foreign", "Home", "1 rue de la Paix", nil, "Paris", "75001", int32(1), nil, nil, authUser).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	resp, err := srv.CreateAddress(context.Background(), &usersGrpc.CreateAddressRequest{
		OwnerType:  usersGrpc.OwnerType_OWNER_TYPE_CLIENT,
		OwnerId:    "c-foreign",
		Name:       "Home",
		Street:     "1 rue de la Paix",
		City:       "Paris",
		ZipCode:    "75001",
		CountryId:  1,
		AuthUserId: authUser,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for foreign owner")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListAddresses_Success(t *testing.T) {
	srv, mock := setupServer(t)

	cols := []string{"id", "owner_type", "owner_id", "name", "street", "additional_street", "city", "zip_code", "country_id", "email", "phone", "archived"}
	mock.ExpectQuery(`SELECT id, owner_type, owner_id`).
		WithArgs("user", authUser, authUser).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(1, "user", authUser, "Home", "1 rue de la Paix", nil, "Paris", "75001", 1, nil, nil, false).
			AddRow(2, "user", authUser, "Office", "5 avenue du Général", nil, "Lyon", "69001", 1, nil, nil, false))

	resp, err := srv.ListAddresses(context.Background(), &usersGrpc.ListAddressesRequest{
		OwnerType:  usersGrpc.OwnerType_OWNER_TYPE_USER,
		OwnerId:    authUser,
		AuthUserId: authUser,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if len(resp.Addresses) != 2 {
		t.Fatalf("expected 2 addresses, got %d", len(resp.Addresses))
	}
	if resp.Addresses[0].OwnerType != usersGrpc.OwnerType_OWNER_TYPE_USER {
		t.Fatalf("expected OwnerType USER, got %v", resp.Addresses[0].OwnerType)
	}
}

func TestArchiveAddress_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE addresses SET archived_at`).
		WithArgs(int32(1), "user", authUser, authUser).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.ArchiveAddress(context.Background(), &usersGrpc.ArchiveAddressRequest{
		AddressId:  1,
		OwnerType:  usersGrpc.OwnerType_OWNER_TYPE_USER,
		OwnerId:    authUser,
		AuthUserId: authUser,
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
		WithArgs(int32(99), "user", authUser, authUser).
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, err := srv.ArchiveAddress(context.Background(), &usersGrpc.ArchiveAddressRequest{
		AddressId:  99,
		OwnerType:  usersGrpc.OwnerType_OWNER_TYPE_USER,
		OwnerId:    authUser,
		AuthUserId: authUser,
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

// TestGetAddress_RejectsForeignOwner: SQL auth predicate filters the row out
// when the supplied owner doesn't trace back to the authenticated user.
// QueryRow returns ErrNoRows → NotFound (NOT InternalError).
func TestGetAddress_RejectsForeignOwner(t *testing.T) {
	srv, mock := setupServer(t)

	cols := []string{"id", "owner_type", "owner_id", "name", "street", "additional_street", "city", "zip_code", "country_id", "email", "phone", "archived"}
	mock.ExpectQuery(`SELECT id, owner_type, owner_id`).
		WithArgs(int32(1), "client", "c-foreign", authUser).
		WillReturnRows(sqlmock.NewRows(cols))

	resp, err := srv.GetAddress(context.Background(), &usersGrpc.GetAddressRequest{
		AddressId:  1,
		OwnerType:  usersGrpc.OwnerType_OWNER_TYPE_CLIENT,
		OwnerId:    "c-foreign",
		AuthUserId: authUser,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for foreign owner")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestListAddresses_FiltersForeignOwner: when the supplied owner can't be
// reached from auth, the auth predicate intersects to zero rows.
func TestListAddresses_FiltersForeignOwner(t *testing.T) {
	srv, mock := setupServer(t)

	cols := []string{"id", "owner_type", "owner_id", "name", "street", "additional_street", "city", "zip_code", "country_id", "email", "phone", "archived"}
	mock.ExpectQuery(`SELECT id, owner_type, owner_id`).
		WithArgs("client", "c-foreign", authUser).
		WillReturnRows(sqlmock.NewRows(cols))

	resp, err := srv.ListAddresses(context.Background(), &usersGrpc.ListAddressesRequest{
		OwnerType:  usersGrpc.OwnerType_OWNER_TYPE_CLIENT,
		OwnerId:    "c-foreign",
		AuthUserId: authUser,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success (empty list), got code %d", resp.Code)
	}
	if len(resp.Addresses) != 0 {
		t.Fatalf("expected 0 addresses, got %d", len(resp.Addresses))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestArchiveAddress_RejectsForeignOwner: distinguishes the auth-blocked case
// from a plain "no such id" — both return NotFound, but the SQL must include
// the auth predicate (sqlmock.WithArgs has 4 args: id, owner_type, owner_id,
// auth_user_id). Removing the predicate would change the arg count and fail.
func TestArchiveAddress_RejectsForeignOwner(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE addresses SET archived_at`).
		WithArgs(int32(1), "client", "c-foreign", authUser).
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, err := srv.ArchiveAddress(context.Background(), &usersGrpc.ArchiveAddressRequest{
		AddressId:  1,
		OwnerType:  usersGrpc.OwnerType_OWNER_TYPE_CLIENT,
		OwnerId:    "c-foreign",
		AuthUserId: authUser,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for foreign owner")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestUpdateAddress_RejectsArchivedClient locks in the secondary auth gate:
// the predicate's inner subquery requires `clients.archived_at IS NULL`. If
// someone removes that filter, an archived client's addresses become
// editable again — silently. This test asserts the SQL still contains the
// `archived_at IS NULL` clause inside the clients subquery.
func TestUpdateAddress_RejectsArchivedClient(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`SELECT client_id FROM clients[\s\S]*archived_at IS NULL`).
		WithArgs("Home", "1 rue de la Paix", nil, "Paris", "75001", int32(1), nil, nil,
			int32(1), "client", "c-archived", authUser).
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, err := srv.UpdateAddress(context.Background(), &usersGrpc.UpdateAddressRequest{
		AddressId:  1,
		OwnerType:  usersGrpc.OwnerType_OWNER_TYPE_CLIENT,
		OwnerId:    "c-archived",
		Name:       "Home",
		Street:     "1 rue de la Paix",
		City:       "Paris",
		ZipCode:    "75001",
		CountryId:  1,
		AuthUserId: authUser,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure when targeting archived client")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestUpdateAddress_RejectsForeignOwner: the SQL auth predicate filters out
// rows whose owner_id doesn't trace back to the authenticated user.
// RowsAffected returns 0 → NotFound.
func TestUpdateAddress_RejectsForeignOwner(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE addresses SET name`).
		WithArgs("Home", "1 rue de la Paix", nil, "Paris", "75001", int32(1), nil, nil,
			int32(1), "client", "c-foreign", authUser).
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, err := srv.UpdateAddress(context.Background(), &usersGrpc.UpdateAddressRequest{
		AddressId:  1,
		OwnerType:  usersGrpc.OwnerType_OWNER_TYPE_CLIENT,
		OwnerId:    "c-foreign",
		Name:       "Home",
		Street:     "1 rue de la Paix",
		City:       "Paris",
		ZipCode:    "75001",
		CountryId:  1,
		AuthUserId: authUser,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for foreign owner")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
