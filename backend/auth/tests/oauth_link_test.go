package tests

import (
	"context"
	"testing"

	"project-devis-auth/actions"
	authGrpc "project-devis-auth/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

// Linking a new identity to the authenticated user: no existing identity → insert.
func TestLinkOAuthIdentity_New(t *testing.T) {
	srv, mock := setupServer(t, &MockUserClient{})

	mock.ExpectQuery(`SELECT user_id FROM oauth_identities WHERE provider = \$1 AND provider_user_id = \$2`).
		WithArgs("google", "sub-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}))

	mock.ExpectExec(`INSERT INTO oauth_identities`).
		WithArgs("google", "sub-1", "user-1", "user@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.LinkOAuthIdentity(context.Background(), &authGrpc.LinkOAuthIdentityRequest{
		UserId:         "user-1",
		Provider:       "google",
		ProviderUserId: "sub-1",
		Email:          "user@example.com",
		EmailVerified:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success || resp.GetCode() != actions.CodeSuccess {
		t.Fatalf("expected success, got success=%v code=%d", resp.Success, resp.GetCode())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// Re-linking the same identity to the same user is idempotent (no insert).
func TestLinkOAuthIdentity_Idempotent(t *testing.T) {
	srv, mock := setupServer(t, &MockUserClient{})

	mock.ExpectQuery(`SELECT user_id FROM oauth_identities WHERE provider = \$1 AND provider_user_id = \$2`).
		WithArgs("google", "sub-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("user-1"))

	resp, err := srv.LinkOAuthIdentity(context.Background(), &authGrpc.LinkOAuthIdentityRequest{
		UserId:         "user-1",
		Provider:       "google",
		ProviderUserId: "sub-1",
		Email:          "user@example.com",
		EmailVerified:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.GetCode())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// Identity already belongs to another account → refuse (anti-takeover).
func TestLinkOAuthIdentity_TakenByAnother(t *testing.T) {
	srv, mock := setupServer(t, &MockUserClient{})

	mock.ExpectQuery(`SELECT user_id FROM oauth_identities WHERE provider = \$1 AND provider_user_id = \$2`).
		WithArgs("google", "sub-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("other-user"))

	resp, err := srv.LinkOAuthIdentity(context.Background(), &authGrpc.LinkOAuthIdentityRequest{
		UserId:         "user-1",
		Provider:       "google",
		ProviderUserId: "sub-1",
		Email:          "user@example.com",
		EmailVerified:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success || resp.GetCode() != actions.CodeOAuthIdentityTaken {
		t.Fatalf("expected CodeOAuthIdentityTaken, got success=%v code=%d", resp.Success, resp.GetCode())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// Unverified email is refused with no DB writes.
func TestLinkOAuthIdentity_EmailNotVerified(t *testing.T) {
	srv, mock := setupServer(t, &MockUserClient{})

	resp, err := srv.LinkOAuthIdentity(context.Background(), &authGrpc.LinkOAuthIdentityRequest{
		UserId:         "user-1",
		Provider:       "google",
		ProviderUserId: "sub-1",
		Email:          "user@example.com",
		EmailVerified:  false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success || resp.GetCode() != actions.CodeOAuthEmailNotVerified {
		t.Fatalf("expected CodeOAuthEmailNotVerified, got success=%v code=%d", resp.Success, resp.GetCode())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// Unlink succeeds when another login method remains (here: a password).
func TestUnlinkOAuthIdentity_OK(t *testing.T) {
	srv, mock := setupServer(t, &MockUserClient{})

	mock.ExpectQuery(`SELECT a.password IS NOT NULL`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"has_password", "count"}).AddRow(true, 1))
	mock.ExpectExec(`DELETE FROM oauth_identities WHERE user_id = \$1 AND provider = \$2`).
		WithArgs("user-1", "google").
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.UnlinkOAuthIdentity(context.Background(), &authGrpc.UnlinkOAuthIdentityRequest{
		UserId:   "user-1",
		Provider: "google",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success || resp.GetCode() != actions.CodeSuccess {
		t.Fatalf("expected success, got success=%v code=%d", resp.Success, resp.GetCode())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// Unlinking the only login method (no password, single identity) is refused.
func TestUnlinkOAuthIdentity_LastMethod(t *testing.T) {
	srv, mock := setupServer(t, &MockUserClient{})

	mock.ExpectQuery(`SELECT a.password IS NOT NULL`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"has_password", "count"}).AddRow(false, 1))

	resp, err := srv.UnlinkOAuthIdentity(context.Background(), &authGrpc.UnlinkOAuthIdentityRequest{
		UserId:   "user-1",
		Provider: "google",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success || resp.GetCode() != actions.CodeLastLoginMethod {
		t.Fatalf("expected CodeLastLoginMethod, got success=%v code=%d", resp.Success, resp.GetCode())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// Listing returns linked providers plus has_password.
func TestListOAuthIdentities_OK(t *testing.T) {
	srv, mock := setupServer(t, &MockUserClient{})

	mock.ExpectQuery(`SELECT password IS NOT NULL FROM auth WHERE user_id = \$1`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"has_password"}).AddRow(true))
	mock.ExpectQuery(`SELECT provider, email FROM oauth_identities WHERE user_id = \$1`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"provider", "email"}).
			AddRow("github", "user@example.com").
			AddRow("google", "user@example.com"))

	resp, err := srv.ListOAuthIdentities(context.Background(), &authGrpc.ListOAuthIdentitiesRequest{
		UserId: "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success || !resp.GetHasPassword() || len(resp.GetIdentities()) != 2 {
		t.Fatalf("unexpected response: success=%v hasPassword=%v count=%d",
			resp.Success, resp.GetHasPassword(), len(resp.GetIdentities()))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
