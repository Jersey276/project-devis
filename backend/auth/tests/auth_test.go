package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"project-devis-auth/actions"
	authGrpc "project-devis-auth/services/grpc"
	userGrpc "project-devis-auth/services/user_auth"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMain(m *testing.M) {
	os.Setenv("APP_KEY", "test-secret-key-for-jwt-signing-32b")
	os.Exit(m.Run())
}

// --- helpers ---

func setupServer(t *testing.T, mockUser userGrpc.UserServiceClient) (*actions.Server, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	srv := actions.NewServerWithClient(db, mockUser)
	return srv, mock
}

// findFieldError reports whether the given field has the given error code in the response.
func findFieldError(fieldErrors []*authGrpc.FormFieldError, field string, code int32) bool {
	for _, fe := range fieldErrors {
		if fe.Field == field {
			for _, c := range fe.ErrorCode {
				if c == code {
					return true
				}
			}
		}
	}
	return false
}

const testUserServiceCodeNotFound int32 = 1001

// --- Register ---

func TestRegister_Success(t *testing.T) {
	mockUser := &MockUserClient{
		CreateUserFn: func(ctx context.Context, req *userGrpc.CreateUserRequest) (*userGrpc.CreateUserResponse, error) {
			return &userGrpc.CreateUserResponse{Success: true, Code: actions.CodeSuccess, UserId: "user-123"}, nil
		},
	}
	srv, mock := setupServer(t, mockUser)

	mock.ExpectQuery(`SELECT email FROM auth WHERE email = \$1`).
		WithArgs("new@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"email"}))

	mock.ExpectBegin()
	mock.ExpectExec(`SELECT pg_advisory_xact_lock\(\$1\)`).
		WithArgs(int64(2026052901)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM auth`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectExec(`INSERT INTO auth`).
		WithArgs("user-123", "new@example.com", sqlmock.AnyArg(), "super_admin", "active", "free").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	resp, err := srv.Register(context.Background(), &authGrpc.RegisterRequest{
		Email:    "new@example.com",
		Password: "Password123!",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Code != actions.CodeSuccess {
		t.Fatalf("expected code %d, got %d", actions.CodeSuccess, resp.Code)
	}
	if len(resp.FieldErrors) != 0 {
		t.Fatalf("expected no field errors on success, got %v", resp.FieldErrors)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	mockUser := &MockUserClient{}
	srv, mock := setupServer(t, mockUser)

	mock.ExpectQuery(`SELECT email FROM auth WHERE email = \$1`).
		WithArgs("existing@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("existing@example.com"))

	resp, err := srv.Register(context.Background(), &authGrpc.RegisterRequest{
		Email:    "existing@example.com",
		Password: "Password123!",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for existing user")
	}
	if resp.Code != actions.CodeUserAlreadyExists {
		t.Fatalf("expected code %d, got %d", actions.CodeUserAlreadyExists, resp.Code)
	}
	if !findFieldError(resp.FieldErrors, "email", actions.FieldErrAlreadyInUse) {
		t.Fatal("expected FieldErrAlreadyInUse on email field")
	}
}

func TestRegister_ValidationInvalidEmail(t *testing.T) {
	mockUser := &MockUserClient{}
	srv, mock := setupServer(t, mockUser)

	resp, err := srv.Register(context.Background(), &authGrpc.RegisterRequest{
		Email:    "not-an-email",
		Password: "Password123!",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for invalid email")
	}
	if !findFieldError(resp.FieldErrors, "email", actions.FieldErrInvalidFormat) {
		t.Fatal("expected FieldErrInvalidFormat on email field")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations (no DB calls expected): %v", err)
	}
}

func TestRegister_ValidationPasswordTooShort(t *testing.T) {
	mockUser := &MockUserClient{}
	srv, mock := setupServer(t, mockUser)

	resp, err := srv.Register(context.Background(), &authGrpc.RegisterRequest{
		Email:    "test@example.com",
		Password: "short",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for short password")
	}
	if !findFieldError(resp.FieldErrors, "password", actions.FieldErrTooShort) {
		t.Fatal("expected FieldErrTooShort on password field")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations (no DB calls expected): %v", err)
	}
}

func TestRegister_ValidationMultipleErrors(t *testing.T) {
	mockUser := &MockUserClient{}
	srv, mock := setupServer(t, mockUser)

	resp, err := srv.Register(context.Background(), &authGrpc.RegisterRequest{
		Email:    "bad-email",
		Password: "x",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for multiple invalid fields")
	}
	if !findFieldError(resp.FieldErrors, "email", actions.FieldErrInvalidFormat) {
		t.Fatal("expected FieldErrInvalidFormat on email field")
	}
	if !findFieldError(resp.FieldErrors, "password", actions.FieldErrTooShort) {
		t.Fatal("expected FieldErrTooShort on password field")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations (no DB calls expected): %v", err)
	}
}

func TestRegister_RollbackOnAuthInsertFailure(t *testing.T) {
	var deletedUserID string
	mockUser := &MockUserClient{
		CreateUserFn: func(ctx context.Context, req *userGrpc.CreateUserRequest) (*userGrpc.CreateUserResponse, error) {
			return &userGrpc.CreateUserResponse{Success: true, Code: actions.CodeSuccess, UserId: "user-456"}, nil
		},
		DeleteUserFn: func(ctx context.Context, req *userGrpc.DeleteUserRequest) (*userGrpc.GenericResponse, error) {
			deletedUserID = req.UserId
			return &userGrpc.GenericResponse{Success: true, Code: actions.CodeSuccess}, nil
		},
	}
	srv, mock := setupServer(t, mockUser)

	mock.ExpectQuery(`SELECT email FROM auth WHERE email = \$1`).
		WithArgs("fail@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"email"}))

	mock.ExpectBegin()
	mock.ExpectExec(`SELECT pg_advisory_xact_lock\(\$1\)`).
		WithArgs(int64(2026052901)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM auth`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectExec(`INSERT INTO auth`).
		WithArgs("user-456", "fail@example.com", sqlmock.AnyArg(), "free_user", "active", "free").
		WillReturnError(sqlmock.ErrCancelled)
	mock.ExpectRollback()

	resp, err := srv.Register(context.Background(), &authGrpc.RegisterRequest{
		Email:    "fail@example.com",
		Password: "Password123!",
	})
	if err == nil {
		t.Fatal("expected error from failed insert")
	}
	if resp.Success {
		t.Fatal("expected failure")
	}
	if resp.Code != actions.CodeInternalError {
		t.Fatalf("expected code %d, got %d", actions.CodeInternalError, resp.Code)
	}
	if deletedUserID != "user-456" {
		t.Fatalf("expected rollback for user-456, got %q", deletedUserID)
	}
}

func TestRegister_UserServiceError(t *testing.T) {
	mockUser := &MockUserClient{
		CreateUserFn: func(ctx context.Context, req *userGrpc.CreateUserRequest) (*userGrpc.CreateUserResponse, error) {
			return &userGrpc.CreateUserResponse{Success: false, Code: actions.CodeUserServiceError}, nil
		},
	}
	srv, mock := setupServer(t, mockUser)

	mock.ExpectQuery(`SELECT email FROM auth WHERE email = \$1`).
		WithArgs("new@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"email"}))

	resp, err := srv.Register(context.Background(), &authGrpc.RegisterRequest{
		Email:    "new@example.com",
		Password: "Password123!",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure when user service errors")
	}
	if resp.Code != actions.CodeUserServiceError {
		t.Fatalf("expected code %d, got %d", actions.CodeUserServiceError, resp.Code)
	}
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	mockUser := &MockUserClient{
		GetUserAccessInfoByEmailFn: func(ctx context.Context, req *userGrpc.GetUserAccessInfoByEmailRequest) (*userGrpc.GetUserAccessInfoResponse, error) {
			return &userGrpc.GetUserAccessInfoResponse{Success: true, Code: 0, UserId: "user-789", Email: req.GetEmail(), Role: "user", Suspended: false}, nil
		},
	}
	srv, mock := setupServer(t, mockUser)

	// bcrypt hash of "password123" (cost 14 is slow for tests, use a pre-computed hash)
	hashedPassword := "$2a$14$XC05j1ejsVEfzQm6g5f52.dw2EN.cadB6VJPH2R5YsdKIpYHmf/NW"

	mock.ExpectQuery(`SELECT password, role, account_status, subscription_tier, session_version FROM auth WHERE user_id = \$1`).
		WithArgs("user-789").
		WillReturnRows(sqlmock.NewRows([]string{"password", "role", "account_status", "subscription_tier", "session_version"}).
			AddRow(hashedPassword, "free_user", "active", "free", 1))

	mock.ExpectExec(`INSERT INTO refresh_tokens`).
		WithArgs("user-789", sqlmock.AnyArg(), sqlmock.AnyArg(), false).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.Login(context.Background(), &authGrpc.LoginRequest{
		Email:    "user@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.GetCode())
	}
	if resp.GetToken() == "" {
		t.Fatal("expected access token, got empty")
	}
	if resp.GetRefreshToken() == "" {
		t.Fatal("expected refresh token, got empty")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	mockUser := &MockUserClient{
		GetUserAccessInfoByEmailFn: func(ctx context.Context, req *userGrpc.GetUserAccessInfoByEmailRequest) (*userGrpc.GetUserAccessInfoResponse, error) {
			return &userGrpc.GetUserAccessInfoResponse{Success: false, Code: testUserServiceCodeNotFound}, nil
		},
	}
	srv, _ := setupServer(t, mockUser)

	resp, err := srv.Login(context.Background(), &authGrpc.LoginRequest{
		Email:    "unknown@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for unknown user")
	}
	if resp.GetCode() != actions.CodeUserNotFound {
		t.Fatalf("expected code %d, got %d", actions.CodeUserNotFound, resp.GetCode())
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	mockUser := &MockUserClient{
		GetUserAccessInfoByEmailFn: func(ctx context.Context, req *userGrpc.GetUserAccessInfoByEmailRequest) (*userGrpc.GetUserAccessInfoResponse, error) {
			return &userGrpc.GetUserAccessInfoResponse{Success: true, Code: 0, UserId: "user-789", Email: req.GetEmail(), Role: "user", Suspended: false}, nil
		},
	}
	srv, mock := setupServer(t, mockUser)

	// Hash for "correctpassword", NOT "wrongpassword"
	hashedPassword := "$2a$14$XC05j1ejsVEfzQm6g5f52.dw2EN.cadB6VJPH2R5YsdKIpYHmf/NW"

	mock.ExpectQuery(`SELECT password, role, account_status, subscription_tier, session_version FROM auth WHERE user_id = \$1`).
		WithArgs("user-789").
		WillReturnRows(sqlmock.NewRows([]string{"password", "role", "account_status", "subscription_tier", "session_version"}).
			AddRow(hashedPassword, "free_user", "active", "free", 1))

	resp, err := srv.Login(context.Background(), &authGrpc.LoginRequest{
		Email:    "user@example.com",
		Password: "wrongpassword",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for wrong password")
	}
	if resp.GetCode() != actions.CodeInvalidCredentials {
		t.Fatalf("expected code %d, got %d", actions.CodeInvalidCredentials, resp.GetCode())
	}
}

func TestLogin_RememberMe(t *testing.T) {
	mockUser := &MockUserClient{
		GetUserAccessInfoByEmailFn: func(ctx context.Context, req *userGrpc.GetUserAccessInfoByEmailRequest) (*userGrpc.GetUserAccessInfoResponse, error) {
			return &userGrpc.GetUserAccessInfoResponse{Success: true, Code: 0, UserId: "user-789", Email: req.GetEmail(), Role: "user", Suspended: false}, nil
		},
	}
	srv, mock := setupServer(t, mockUser)

	hashedPassword := "$2a$14$XC05j1ejsVEfzQm6g5f52.dw2EN.cadB6VJPH2R5YsdKIpYHmf/NW"

	mock.ExpectQuery(`SELECT password, role, account_status, subscription_tier, session_version FROM auth WHERE user_id = \$1`).
		WithArgs("user-789").
		WillReturnRows(sqlmock.NewRows([]string{"password", "role", "account_status", "subscription_tier", "session_version"}).
			AddRow(hashedPassword, "free_user", "active", "free", 1))

	mock.ExpectExec(`INSERT INTO refresh_tokens`).
		WithArgs("user-789", sqlmock.AnyArg(), sqlmock.AnyArg(), true).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.Login(context.Background(), &authGrpc.LoginRequest{
		Email:      "user@example.com",
		Password:   "password123",
		RememberMe: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.GetCode())
	}
	if resp.GetToken() == "" || resp.GetRefreshToken() == "" {
		t.Fatal("expected both tokens")
	}
}

// --- RefreshToken ---

func TestRefreshToken_Success(t *testing.T) {
	mockUser := &MockUserClient{
		GetUserAccessInfoFn: func(ctx context.Context, req *userGrpc.GetUserAccessInfoRequest) (*userGrpc.GetUserAccessInfoResponse, error) {
			return &userGrpc.GetUserAccessInfoResponse{Success: true, Code: 0, UserId: "user-789", Email: "user@example.com", Role: "user", Suspended: false}, nil
		},
	}
	srv, mock := setupServer(t, mockUser)

	fakeTokenHash := sqlmock.AnyArg()

	mock.ExpectQuery(`SELECT user_id, expires_at, remember_me FROM refresh_tokens WHERE token_hash = \$1`).
		WithArgs(fakeTokenHash).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at", "remember_me"}).
			AddRow("user-789", time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC), true))

	mock.ExpectQuery(`SELECT role, account_status, subscription_tier, session_version FROM auth WHERE user_id = \$1`).
		WithArgs("user-789").
		WillReturnRows(sqlmock.NewRows([]string{"role", "account_status", "subscription_tier", "session_version"}).AddRow("free_user", "active", "free", 1))
	mock.ExpectExec(`DELETE FROM refresh_tokens WHERE token_hash = \$1`).
		WithArgs(fakeTokenHash).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`INSERT INTO refresh_tokens`).
		WithArgs("user-789", sqlmock.AnyArg(), sqlmock.AnyArg(), true).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.RefreshToken(context.Background(), &authGrpc.RefreshTokenRequest{
		RefreshToken: "some-uuid-token",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.GetCode())
	}
	if resp.GetToken() == "" {
		t.Fatal("expected new access token")
	}
	if resp.GetRefreshToken() == "" {
		t.Fatal("expected new refresh token")
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	mockUser := &MockUserClient{}
	srv, mock := setupServer(t, mockUser)

	mock.ExpectQuery(`SELECT user_id, expires_at, remember_me FROM refresh_tokens WHERE token_hash = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at", "remember_me"}))

	resp, err := srv.RefreshToken(context.Background(), &authGrpc.RefreshTokenRequest{
		RefreshToken: "invalid-token",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for invalid refresh token")
	}
	if resp.GetCode() != actions.CodeInvalidRefreshToken {
		t.Fatalf("expected code %d, got %d", actions.CodeInvalidRefreshToken, resp.GetCode())
	}
}

func TestRefreshToken_SuspendedAccount_ReturnsInvalidRefreshToken(t *testing.T) {
	mockUser := &MockUserClient{
		GetUserAccessInfoFn: func(ctx context.Context, req *userGrpc.GetUserAccessInfoRequest) (*userGrpc.GetUserAccessInfoResponse, error) {
			return &userGrpc.GetUserAccessInfoResponse{Success: true, Code: 0, UserId: "user-789", Email: "user@example.com", Role: "user", Suspended: true}, nil
		},
	}
	srv, mock := setupServer(t, mockUser)

	mock.ExpectQuery(`SELECT user_id, expires_at, remember_me FROM refresh_tokens WHERE token_hash = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at", "remember_me"}).
			AddRow("user-789", time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC), true))

	mock.ExpectExec(`DELETE FROM refresh_tokens WHERE token_hash = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.RefreshToken(context.Background(), &authGrpc.RefreshTokenRequest{RefreshToken: "some-uuid-token"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for suspended user")
	}
	if resp.GetCode() != actions.CodeInvalidRefreshToken {
		t.Fatalf("expected code %d, got %d", actions.CodeInvalidRefreshToken, resp.GetCode())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// --- Logout ---

func TestLogout_Success(t *testing.T) {
	mockUser := &MockUserClient{}
	srv, mock := setupServer(t, mockUser)

	mock.ExpectExec(`DELETE FROM refresh_tokens WHERE token_hash = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.Logout(context.Background(), &authGrpc.LogoutRequest{
		RefreshToken: "some-uuid-token",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

// --- ResetPassword ---

func TestResetPassword_UserNotFound_ReturnsGenericSuccess(t *testing.T) {
	mockUser := &MockUserClient{
		GetUserAccessInfoByEmailFn: func(ctx context.Context, req *userGrpc.GetUserAccessInfoByEmailRequest) (*userGrpc.GetUserAccessInfoResponse, error) {
			return &userGrpc.GetUserAccessInfoResponse{Success: false, Code: testUserServiceCodeNotFound}, nil
		},
	}
	srv, mock := setupServer(t, mockUser)

	resp, err := srv.ResetPassword(context.Background(), &authGrpc.ResetPasswordRequest{
		Email: "missing@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success for anti-enumeration, got code %d", resp.Code)
	}
	if resp.Code != actions.CodeSuccess {
		t.Fatalf("expected code %d, got %d", actions.CodeSuccess, resp.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestResetPassword_ExistingUser_CreatesResetToken(t *testing.T) {
	mockUser := &MockUserClient{
		GetUserAccessInfoByEmailFn: func(ctx context.Context, req *userGrpc.GetUserAccessInfoByEmailRequest) (*userGrpc.GetUserAccessInfoResponse, error) {
			return &userGrpc.GetUserAccessInfoResponse{Success: true, Code: 0, UserId: "user-123", Email: req.GetEmail(), Role: "user", Suspended: false}, nil
		},
	}
	srv, mock := setupServer(t, mockUser)

	mock.ExpectExec(`INSERT INTO password_reset_tokens \(user_id, token_hash, expires_at\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs("user-123", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.ResetPassword(context.Background(), &authGrpc.ResetPasswordRequest{
		Email: "known@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Code != actions.CodeSuccess {
		t.Fatalf("expected code %d, got %d", actions.CodeSuccess, resp.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// --- ConfirmResetPassword ---

func TestConfirmResetPassword_Success_UpdatesPasswordRevokesSessionsAndConsumesToken(t *testing.T) {
	mockUser := &MockUserClient{}
	srv, mock := setupServer(t, mockUser)

	mock.ExpectQuery(`SELECT user_id, expires_at, used_at FROM password_reset_tokens WHERE token_hash = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at", "used_at"}).
			AddRow("user-abc", time.Now().Add(10*time.Minute), nil))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE auth SET password = \$1, session_version = session_version \+ 1 WHERE user_id = \$2`).
		WithArgs(sqlmock.AnyArg(), "user-abc").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM refresh_tokens WHERE user_id = \$1`).
		WithArgs("user-abc").
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec(`UPDATE password_reset_tokens SET used_at = NOW\(\) WHERE token_hash = \$1 AND used_at IS NULL AND expires_at > NOW\(\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	resp, err := srv.ConfirmResetPassword(context.Background(), &authGrpc.ConfirmResetPasswordRequest{
		Token:       "reset-token",
		NewPassword: "StrongPass123!",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Code != actions.CodeSuccess {
		t.Fatalf("expected code %d, got %d", actions.CodeSuccess, resp.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestConfirmResetPassword_InvalidToken_ReturnsBusinessError(t *testing.T) {
	mockUser := &MockUserClient{}
	srv, mock := setupServer(t, mockUser)

	mock.ExpectQuery(`SELECT user_id, expires_at, used_at FROM password_reset_tokens WHERE token_hash = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at", "used_at"}))

	resp, err := srv.ConfirmResetPassword(context.Background(), &authGrpc.ConfirmResetPasswordRequest{
		Token:       "missing-token",
		NewPassword: "StrongPass123!",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure")
	}
	if resp.Code != actions.CodeInvalidResetToken {
		t.Fatalf("expected code %d, got %d", actions.CodeInvalidResetToken, resp.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestConfirmResetPassword_ExpiredToken_ReturnsExpiredCode(t *testing.T) {
	mockUser := &MockUserClient{}
	srv, mock := setupServer(t, mockUser)

	mock.ExpectQuery(`SELECT user_id, expires_at, used_at FROM password_reset_tokens WHERE token_hash = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at", "used_at"}).
			AddRow("user-abc", time.Now().Add(-1*time.Minute), nil))

	resp, err := srv.ConfirmResetPassword(context.Background(), &authGrpc.ConfirmResetPasswordRequest{
		Token:       "expired-token",
		NewPassword: "StrongPass123!",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure")
	}
	if resp.Code != actions.CodeExpiredResetToken {
		t.Fatalf("expected code %d, got %d", actions.CodeExpiredResetToken, resp.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestConfirmResetPassword_WeakPassword_ReturnsWeakPasswordCode(t *testing.T) {
	mockUser := &MockUserClient{}
	srv, mock := setupServer(t, mockUser)

	resp, err := srv.ConfirmResetPassword(context.Background(), &authGrpc.ConfirmResetPasswordRequest{
		Token:       "any-token",
		NewPassword: "short",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure")
	}
	if resp.Code != actions.CodeWeakPassword {
		t.Fatalf("expected code %d, got %d", actions.CodeWeakPassword, resp.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations (no DB calls expected): %v", err)
	}
}

// --- UpdatePassword ---

func TestUpdatePassword_Success_UpdatesPasswordAndRevokesOtherSessions(t *testing.T) {
	mockUser := &MockUserClient{
		GetUserAccessInfoByEmailFn: func(ctx context.Context, req *userGrpc.GetUserAccessInfoByEmailRequest) (*userGrpc.GetUserAccessInfoResponse, error) {
			return &userGrpc.GetUserAccessInfoResponse{Success: true, Code: 0, UserId: "user-123", Email: req.GetEmail(), Role: "user", Suspended: false}, nil
		},
	}
	srv, mock := setupServer(t, mockUser)

	oldPasswordHash := "$2a$14$XC05j1ejsVEfzQm6g5f52.dw2EN.cadB6VJPH2R5YsdKIpYHmf/NW" // password123

	mock.ExpectQuery(`SELECT password FROM auth WHERE user_id = \$1`).
		WithArgs("user-123").
		WillReturnRows(sqlmock.NewRows([]string{"password"}).AddRow(oldPasswordHash))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE auth SET password = \$1, session_version = session_version \+ 1 WHERE user_id = \$2`).
		WithArgs(sqlmock.AnyArg(), "user-123").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM refresh_tokens WHERE user_id = \$1 AND token_hash <> \$2`).
		WithArgs("user-123", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	resp, err := srv.UpdatePassword(context.Background(), &authGrpc.UpdatePasswordRequest{
		Email:               "user@example.com",
		OldPassword:         "password123",
		NewPassword:         "StrongPass123!",
		CurrentRefreshToken: "current-refresh-token",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Code != actions.CodeSuccess {
		t.Fatalf("expected code %d, got %d", actions.CodeSuccess, resp.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdatePassword_InvalidOldPassword_ReturnsInvalidCredentials(t *testing.T) {
	mockUser := &MockUserClient{
		GetUserAccessInfoByEmailFn: func(ctx context.Context, req *userGrpc.GetUserAccessInfoByEmailRequest) (*userGrpc.GetUserAccessInfoResponse, error) {
			return &userGrpc.GetUserAccessInfoResponse{Success: true, Code: 0, UserId: "user-123", Email: req.GetEmail(), Role: "user", Suspended: false}, nil
		},
	}
	_, mock := setupServer(t, mockUser)

	oldPasswordHash := "$2a$14$XC05j1ejsVEfzQm6g5f52.dw2EN.cadB6VJPH2R5YsdKIpYHmf/NW" // password123

	mock.ExpectQuery(`SELECT password FROM auth WHERE user_id = \$1`).
		WithArgs("user-123").
		WillReturnRows(sqlmock.NewRows([]string{"password"}).AddRow(oldPasswordHash))
}