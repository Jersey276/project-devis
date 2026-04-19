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

	mock.ExpectExec(`INSERT INTO auth`).
		WithArgs("user-123", "new@example.com", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.Register(context.Background(), &authGrpc.RegisterRequest{
		Name:     "testuser",
		Email:    "new@example.com",
		Password: "password123",
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
		Name:     "testuser",
		Email:    "existing@example.com",
		Password: "password123",
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

func TestRegister_ValidationMissingName(t *testing.T) {
	mockUser := &MockUserClient{}
	srv, mock := setupServer(t, mockUser)

	resp, err := srv.Register(context.Background(), &authGrpc.RegisterRequest{
		Name:     "",
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for missing name")
	}
	if !findFieldError(resp.FieldErrors, "name", actions.FieldErrRequired) {
		t.Fatal("expected FieldErrRequired on name field")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations (no DB calls expected): %v", err)
	}
}

func TestRegister_ValidationInvalidEmail(t *testing.T) {
	mockUser := &MockUserClient{}
	srv, mock := setupServer(t, mockUser)

	resp, err := srv.Register(context.Background(), &authGrpc.RegisterRequest{
		Name:     "testuser",
		Email:    "not-an-email",
		Password: "password123",
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
		Name:     "testuser",
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
		Name:     "",
		Email:    "bad-email",
		Password: "x",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for multiple invalid fields")
	}
	if !findFieldError(resp.FieldErrors, "name", actions.FieldErrRequired) {
		t.Fatal("expected FieldErrRequired on name field")
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

	mock.ExpectExec(`INSERT INTO auth`).
		WithArgs("user-456", "fail@example.com", sqlmock.AnyArg()).
		WillReturnError(sqlmock.ErrCancelled)

	resp, err := srv.Register(context.Background(), &authGrpc.RegisterRequest{
		Name:     "testuser",
		Email:    "fail@example.com",
		Password: "password123",
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
		Name:     "testuser",
		Email:    "new@example.com",
		Password: "password123",
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
	mockUser := &MockUserClient{}
	srv, mock := setupServer(t, mockUser)

	// bcrypt hash of "password123" (cost 14 is slow for tests, use a pre-computed hash)
	hashedPassword := "$2a$14$XC05j1ejsVEfzQm6g5f52.dw2EN.cadB6VJPH2R5YsdKIpYHmf/NW"

	mock.ExpectQuery(`SELECT email, password, user_id FROM auth WHERE email = \$1`).
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"email", "password", "user_id"}).
			AddRow("user@example.com", hashedPassword, "user-789"))

	// INSERT refresh token
	mock.ExpectExec(`INSERT INTO refresh_tokens`).
		WithArgs("user-789", sqlmock.AnyArg(), sqlmock.AnyArg()).
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
	mockUser := &MockUserClient{}
	srv, mock := setupServer(t, mockUser)

	mock.ExpectQuery(`SELECT email, password, user_id FROM auth WHERE email = \$1`).
		WithArgs("unknown@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"email", "password", "user_id"}))

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
	mockUser := &MockUserClient{}
	srv, mock := setupServer(t, mockUser)

	// Hash for "correctpassword", NOT "wrongpassword"
	hashedPassword := "$2a$14$XC05j1ejsVEfzQm6g5f52.dw2EN.cadB6VJPH2R5YsdKIpYHmf/NW"

	mock.ExpectQuery(`SELECT email, password, user_id FROM auth WHERE email = \$1`).
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"email", "password", "user_id"}).
			AddRow("user@example.com", hashedPassword, "user-789"))

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
	mockUser := &MockUserClient{}
	srv, mock := setupServer(t, mockUser)

	hashedPassword := "$2a$14$XC05j1ejsVEfzQm6g5f52.dw2EN.cadB6VJPH2R5YsdKIpYHmf/NW"

	mock.ExpectQuery(`SELECT email, password, user_id FROM auth WHERE email = \$1`).
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"email", "password", "user_id"}).
			AddRow("user@example.com", hashedPassword, "user-789"))

	mock.ExpectExec(`INSERT INTO refresh_tokens`).
		WithArgs("user-789", sqlmock.AnyArg(), sqlmock.AnyArg()).
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
	mockUser := &MockUserClient{}
	srv, mock := setupServer(t, mockUser)

	fakeTokenHash := sqlmock.AnyArg()

	mock.ExpectQuery(`SELECT user_id, expires_at FROM refresh_tokens WHERE token_hash = \$1`).
		WithArgs(fakeTokenHash).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at"}).
			AddRow("user-789", time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)))

	mock.ExpectQuery(`SELECT email FROM auth WHERE user_id = \$1`).
		WithArgs("user-789").
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("user@example.com"))

	mock.ExpectExec(`DELETE FROM refresh_tokens WHERE token_hash = \$1`).
		WithArgs(fakeTokenHash).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`INSERT INTO refresh_tokens`).
		WithArgs("user-789", sqlmock.AnyArg(), sqlmock.AnyArg()).
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

	mock.ExpectQuery(`SELECT user_id, expires_at FROM refresh_tokens WHERE token_hash = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at"}))

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
