package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	auth "gateway/auth"
	"gateway/authcookie"
	"gateway/middleware"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type mockAuthClient struct {
	resetFn        func(context.Context, *auth.ResetPasswordRequest) (*auth.GenericResponse, error)
	confirmFn      func(context.Context, *auth.ConfirmResetPasswordRequest) (*auth.GenericResponse, error)
	updateFn       func(context.Context, *auth.UpdatePasswordRequest) (*auth.GenericResponse, error)
	resendVerifyFn func(context.Context, *auth.ResendEmailVerificationRequest) (*auth.GenericResponse, error)
}

func (m *mockAuthClient) Register(context.Context, *auth.RegisterRequest, ...grpc.CallOption) (*auth.FormGenericResponse, error) {
	return &auth.FormGenericResponse{Success: true, Code: CodeSuccess}, nil
}

func (m *mockAuthClient) Login(context.Context, *auth.LoginRequest, ...grpc.CallOption) (*auth.LoginResponse, error) {
	token := "token"
	refresh := "refresh"
	return &auth.LoginResponse{Success: true, Token: &token, RefreshToken: &refresh}, nil
}

func (m *mockAuthClient) ResetPassword(ctx context.Context, req *auth.ResetPasswordRequest, _ ...grpc.CallOption) (*auth.GenericResponse, error) {
	if m.resetFn != nil {
		return m.resetFn(ctx, req)
	}
	return &auth.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

func (m *mockAuthClient) ConfirmResetPassword(ctx context.Context, req *auth.ConfirmResetPasswordRequest, _ ...grpc.CallOption) (*auth.GenericResponse, error) {
	if m.confirmFn != nil {
		return m.confirmFn(ctx, req)
	}
	return &auth.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

func (m *mockAuthClient) UpdatePassword(ctx context.Context, req *auth.UpdatePasswordRequest, _ ...grpc.CallOption) (*auth.GenericResponse, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, req)
	}
	return &auth.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

func (m *mockAuthClient) VerifyEmail(context.Context, *auth.VerifyEmailRequest, ...grpc.CallOption) (*auth.GenericResponse, error) {
	return &auth.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

func (m *mockAuthClient) RefreshToken(context.Context, *auth.RefreshTokenRequest, ...grpc.CallOption) (*auth.LoginResponse, error) {
	token := "token"
	refresh := "refresh"
	return &auth.LoginResponse{Success: true, Token: &token, RefreshToken: &refresh}, nil
}

func (m *mockAuthClient) Logout(context.Context, *auth.LogoutRequest, ...grpc.CallOption) (*auth.GenericResponse, error) {
	return &auth.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

func (m *mockAuthClient) IntrospectToken(context.Context, *auth.IntrospectTokenRequest, ...grpc.CallOption) (*auth.IntrospectTokenResponse, error) {
	return &auth.IntrospectTokenResponse{
		Success: true,
		Code:    CodeSuccess,
		Context: &auth.AccessContext{
			UserId:           "user-1",
			Email:            "user@example.com",
			Role:             "free_user",
			AccountStatus:    "active",
			SubscriptionTier: "free",
			SessionVersion:   1,
		},
	}, nil
}

func (m *mockAuthClient) UpdateSubscriptionTier(context.Context, *auth.UpdateSubscriptionTierRequest, ...grpc.CallOption) (*auth.GenericResponse, error) {
	return &auth.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

func (m *mockAuthClient) UpdateRole(context.Context, *auth.UpdateRoleRequest, ...grpc.CallOption) (*auth.GenericResponse, error) {
	return &auth.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

func (m *mockAuthClient) ResendEmailVerification(ctx context.Context, req *auth.ResendEmailVerificationRequest, _ ...grpc.CallOption) (*auth.GenericResponse, error) {
	if m.resendVerifyFn != nil {
		return m.resendVerifyFn(ctx, req)
	}
	return &auth.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

func resetAuthLimiterStateForTests() {
	resetPasswordIPLimiter = newSlidingWindowLimiter()
	resetPasswordEmailLimiter = newSlidingWindowLimiter()
	confirmResetIPLimiter = newSlidingWindowLimiter()
	resendVerificationLimiter = newSlidingWindowLimiter()
}

func newJSONContext(method, path, body, remoteAddr string) (*gin.Context, *httptest.ResponseRecorder) {
	res := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(res)
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = remoteAddr
	c.Request = req
	return c, res
}

func decodeJSONBody(t *testing.T, res *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return body
}

func TestResetPassword_RateLimitedByEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAuthLimiterStateForTests()
	client := &mockAuthClient{}

	for i := range 3 {
		ctx, res := newJSONContext(http.MethodPost, "/api/auth/password/reset", `{"email":"same@example.com"}`, "10.0.0.1:1234")
		ResetPassword(ctx, client)
		if res.Code != http.StatusOK {
			t.Fatalf("attempt %d expected 200, got %d", i+1, res.Code)
		}
	}

	ctx, res := newJSONContext(http.MethodPost, "/api/auth/password/reset", `{"email":"same@example.com"}`, "10.0.0.1:1234")
	ResetPassword(ctx, client)
	if res.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for email rate limit, got %d", res.Code)
	}
}

func TestConfirmResetPassword_MapsInvalidResetTokenToBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAuthLimiterStateForTests()
	client := &mockAuthClient{
		confirmFn: func(context.Context, *auth.ConfirmResetPasswordRequest) (*auth.GenericResponse, error) {
			return &auth.GenericResponse{Success: false, Code: CodeInvalidResetToken}, nil
		},
	}

	ctx, res := newJSONContext(http.MethodPost, "/api/auth/password/confirm-reset", `{"token":"bad","new_password":"StrongPass123!"}`, "10.0.0.2:1234")
	ConfirmResetPassword(ctx, client)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
	body := decodeJSONBody(t, res)
	if code, ok := body["code"].(float64); !ok || int32(code) != CodeInvalidResetToken {
		t.Fatalf("expected code %d, got %v", CodeInvalidResetToken, body["code"])
	}
}

func TestConfirmResetPassword_MapsExpiredResetTokenToGone(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAuthLimiterStateForTests()
	client := &mockAuthClient{
		confirmFn: func(context.Context, *auth.ConfirmResetPasswordRequest) (*auth.GenericResponse, error) {
			return &auth.GenericResponse{Success: false, Code: CodeExpiredResetToken}, nil
		},
	}

	ctx, res := newJSONContext(http.MethodPost, "/api/auth/password/confirm-reset", `{"token":"expired","new_password":"StrongPass123!"}`, "10.0.0.3:1234")
	ConfirmResetPassword(ctx, client)

	if res.Code != http.StatusGone {
		t.Fatalf("expected 410, got %d", res.Code)
	}
	body := decodeJSONBody(t, res)
	if code, ok := body["code"].(float64); !ok || int32(code) != CodeExpiredResetToken {
		t.Fatalf("expected code %d, got %v", CodeExpiredResetToken, body["code"])
	}
}

func TestConfirmResetPassword_RateLimitedByIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAuthLimiterStateForTests()
	client := &mockAuthClient{}

	for i := range 10 {
		ctx, res := newJSONContext(http.MethodPost, "/api/auth/password/confirm-reset", `{"token":"ok","new_password":"StrongPass123!"}`, "10.0.0.4:1234")
		ConfirmResetPassword(ctx, client)
		if res.Code != http.StatusOK {
			t.Fatalf("attempt %d expected 200, got %d", i+1, res.Code)
		}
	}

	ctx, res := newJSONContext(http.MethodPost, "/api/auth/password/confirm-reset", `{"token":"ok","new_password":"StrongPass123!"}`, "10.0.0.4:1234")
	ConfirmResetPassword(ctx, client)
	if res.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for confirm-reset rate limit, got %d", res.Code)
	}
}

func TestUpdatePassword_UsesEmailFromAuthContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := &mockAuthClient{}
	forwardedEmail := ""
	forwardedRefreshToken := ""
	client.updateFn = func(_ context.Context, req *auth.UpdatePasswordRequest) (*auth.GenericResponse, error) {
		forwardedEmail = req.Email
		forwardedRefreshToken = req.CurrentRefreshToken
		return &auth.GenericResponse{Success: true, Code: CodeSuccess}, nil
	}

	ctx, res := newJSONContext(http.MethodPost, "/api/auth/password/update", `{"email":"attacker@example.com","old_password":"old","new_password":"StrongPass123!"}`, "10.0.0.5:1234")
	ctx.Set(middleware.CtxEmail, "owner@example.com")
	ctx.Request.AddCookie(&http.Cookie{Name: authcookie.RefreshName, Value: "current-refresh-token"})
	UpdatePassword(ctx, client)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	if forwardedEmail != "owner@example.com" {
		t.Fatalf("expected forwarded email owner@example.com, got %q", forwardedEmail)
	}
	if forwardedRefreshToken != "current-refresh-token" {
		t.Fatalf("expected forwarded refresh token current-refresh-token, got %q", forwardedRefreshToken)
	}
}

func TestUpdatePassword_RequiresAuthContextEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := &mockAuthClient{}

	ctx, res := newJSONContext(http.MethodPost, "/api/auth/password/update", `{"old_password":"old","new_password":"StrongPass123!"}`, "10.0.0.6:1234")
	UpdatePassword(ctx, client)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.Code)
	}
}

func TestResendEmailVerification_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAuthLimiterStateForTests()
	client := &mockAuthClient{}

	ctx, res := newJSONContext(http.MethodPost, "/api/auth/email/resend-verification", `{}`, "10.0.0.7:1234")
	ctx.Set(middleware.CtxUserID, "user-42")
	ResendEmailVerification(ctx, client)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}

func TestResendEmailVerification_AlreadyVerified(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAuthLimiterStateForTests()
	client := &mockAuthClient{
		resendVerifyFn: func(_ context.Context, _ *auth.ResendEmailVerificationRequest) (*auth.GenericResponse, error) {
			return &auth.GenericResponse{Success: false, Code: CodeAlreadyVerified}, nil
		},
	}

	ctx, res := newJSONContext(http.MethodPost, "/api/auth/email/resend-verification", `{}`, "10.0.0.8:1234")
	ctx.Set(middleware.CtxUserID, "user-42")
	ResendEmailVerification(ctx, client)

	if res.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", res.Code)
	}
}

func TestResendEmailVerification_RateLimited(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAuthLimiterStateForTests()
	client := &mockAuthClient{}

	for i := range 3 {
		ctx, res := newJSONContext(http.MethodPost, "/api/auth/email/resend-verification", `{}`, "10.0.0.9:1234")
		ctx.Set(middleware.CtxUserID, "user-rl")
		ResendEmailVerification(ctx, client)
		if res.Code != http.StatusOK {
			t.Fatalf("attempt %d expected 200, got %d", i+1, res.Code)
		}
	}

	ctx, res := newJSONContext(http.MethodPost, "/api/auth/email/resend-verification", `{}`, "10.0.0.9:1234")
	ctx.Set(middleware.CtxUserID, "user-rl")
	ResendEmailVerification(ctx, client)
	if res.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after rate limit, got %d", res.Code)
	}
}

func TestResendEmailVerification_RequiresAuthContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := &mockAuthClient{}

	ctx, res := newJSONContext(http.MethodPost, "/api/auth/email/resend-verification", `{}`, "10.0.0.10:1234")
	// No user_id set in context
	ResendEmailVerification(ctx, client)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.Code)
	}
}
