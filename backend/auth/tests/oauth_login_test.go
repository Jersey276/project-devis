package tests

import (
	"context"
	"testing"

	"project-devis-auth/actions"
	authGrpc "project-devis-auth/services/grpc"
	userGrpc "project-devis-auth/services/user_auth"

	"github.com/DATA-DOG/go-sqlmock"
)

// (1) Returning OAuth user: identity already exists → straight login, no CreateUser.
func TestOAuthLogin_ExistingIdentity(t *testing.T) {
	createCalled := false
	mockUser := &MockUserClient{
		CreateUserFn: func(ctx context.Context, req *userGrpc.CreateUserRequest) (*userGrpc.CreateUserResponse, error) {
			createCalled = true
			return &userGrpc.CreateUserResponse{Success: true, UserId: "user-x"}, nil
		},
	}
	srv, mock := setupServer(t, mockUser)

	mock.ExpectQuery(`SELECT user_id FROM oauth_identities WHERE provider = \$1 AND provider_user_id = \$2`).
		WithArgs("google", "sub-123").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("user-789"))

	mock.ExpectQuery(`SELECT role, account_status, subscription_tier, session_version FROM auth WHERE user_id = \$1`).
		WithArgs("user-789").
		WillReturnRows(sqlmock.NewRows([]string{"role", "account_status", "subscription_tier", "session_version"}).
			AddRow("free_user", "active", "free", 1))

	mock.ExpectExec(`INSERT INTO refresh_tokens`).
		WithArgs("user-789", sqlmock.AnyArg(), sqlmock.AnyArg(), false).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.OAuthLogin(context.Background(), &authGrpc.OAuthLoginRequest{
		Provider:       "google",
		ProviderUserId: "sub-123",
		Email:          "user@example.com",
		EmailVerified:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success || resp.GetToken() == "" || resp.GetRefreshToken() == "" {
		t.Fatalf("expected successful login with tokens, got code %d", resp.GetCode())
	}
	if createCalled {
		t.Fatal("CreateUser must not be called for an existing identity")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// (2) Link by email: no identity, but an account with the same email exists →
// insert identity, mark verified, login. CreateUser must not be called.
func TestOAuthLogin_LinkByEmail(t *testing.T) {
	createCalled := false
	mockUser := &MockUserClient{
		CreateUserFn: func(ctx context.Context, req *userGrpc.CreateUserRequest) (*userGrpc.CreateUserResponse, error) {
			createCalled = true
			return &userGrpc.CreateUserResponse{Success: true, UserId: "user-x"}, nil
		},
	}
	srv, mock := setupServer(t, mockUser)

	mock.ExpectQuery(`SELECT user_id FROM oauth_identities WHERE provider = \$1 AND provider_user_id = \$2`).
		WithArgs("github", "gh-42").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}))

	mock.ExpectQuery(`SELECT user_id FROM auth WHERE email = \$1`).
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("user-555"))

	mock.ExpectExec(`INSERT INTO oauth_identities`).
		WithArgs("github", "gh-42", "user-555", "user@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`UPDATE auth SET email_verified = true WHERE user_id = \$1`).
		WithArgs("user-555").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(`SELECT role, account_status, subscription_tier, session_version FROM auth WHERE user_id = \$1`).
		WithArgs("user-555").
		WillReturnRows(sqlmock.NewRows([]string{"role", "account_status", "subscription_tier", "session_version"}).
			AddRow("free_user", "active", "free", 1))

	mock.ExpectExec(`INSERT INTO refresh_tokens`).
		WithArgs("user-555", sqlmock.AnyArg(), sqlmock.AnyArg(), false).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.OAuthLogin(context.Background(), &authGrpc.OAuthLoginRequest{
		Provider:       "github",
		ProviderUserId: "gh-42",
		Email:          "user@example.com",
		EmailVerified:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.GetCode())
	}
	if createCalled {
		t.Fatal("CreateUser must not be called when linking by email")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// (3) Provision a brand-new user (first account → super_admin). CreateUser called.
func TestOAuthLogin_ProvisionNewFirstUser(t *testing.T) {
	createCalled := false
	mockUser := &MockUserClient{
		CreateUserFn: func(ctx context.Context, req *userGrpc.CreateUserRequest) (*userGrpc.CreateUserResponse, error) {
			createCalled = true
			if !req.GetIsAdmin() {
				t.Error("expected IsAdmin=true for the first user")
			}
			return &userGrpc.CreateUserResponse{Success: true, UserId: "user-new"}, nil
		},
	}
	srv, mock := setupServer(t, mockUser)

	mock.ExpectQuery(`SELECT user_id FROM oauth_identities WHERE provider = \$1 AND provider_user_id = \$2`).
		WithArgs("microsoft", "ms-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}))

	mock.ExpectQuery(`SELECT user_id FROM auth WHERE email = \$1`).
		WithArgs("new@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}))

	// provisionUser: pre-count (0) → first user.
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM auth`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectBegin()
	mock.ExpectExec(`SELECT pg_advisory_xact_lock\(\$1\)`).
		WithArgs(int64(2026052901)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM auth`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO auth`).
		WithArgs("user-new", "new@example.com", sqlmock.AnyArg(), "super_admin", "active", "free", true).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mock.ExpectExec(`INSERT INTO oauth_identities`).
		WithArgs("microsoft", "ms-1", "user-new", "new@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(`SELECT role, account_status, subscription_tier, session_version FROM auth WHERE user_id = \$1`).
		WithArgs("user-new").
		WillReturnRows(sqlmock.NewRows([]string{"role", "account_status", "subscription_tier", "session_version"}).
			AddRow("super_admin", "active", "free", 1))
	mock.ExpectExec(`INSERT INTO refresh_tokens`).
		WithArgs("user-new", sqlmock.AnyArg(), sqlmock.AnyArg(), false).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.OAuthLogin(context.Background(), &authGrpc.OAuthLoginRequest{
		Provider:       "microsoft",
		ProviderUserId: "ms-1",
		Email:          "new@example.com",
		EmailVerified:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.GetCode())
	}
	if !createCalled {
		t.Fatal("CreateUser must be called when provisioning a new user")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// (4) Unverified email is refused with no DB writes.
func TestOAuthLogin_EmailNotVerified(t *testing.T) {
	mockUser := &MockUserClient{}
	srv, mock := setupServer(t, mockUser)

	resp, err := srv.OAuthLogin(context.Background(), &authGrpc.OAuthLoginRequest{
		Provider:       "google",
		ProviderUserId: "sub-1",
		Email:          "user@example.com",
		EmailVerified:  false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for unverified email")
	}
	if resp.GetCode() != actions.CodeOAuthEmailNotVerified {
		t.Fatalf("expected code %d, got %d", actions.CodeOAuthEmailNotVerified, resp.GetCode())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations (no DB calls expected): %v", err)
	}
}

// Invalid provider / empty sub is rejected up front.
func TestOAuthLogin_InvalidProvider(t *testing.T) {
	srv, mock := setupServer(t, &MockUserClient{})

	resp, err := srv.OAuthLogin(context.Background(), &authGrpc.OAuthLoginRequest{
		Provider:       "myspace",
		ProviderUserId: "x",
		Email:          "user@example.com",
		EmailVerified:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success || resp.GetCode() != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got success=%v code=%d", resp.Success, resp.GetCode())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
