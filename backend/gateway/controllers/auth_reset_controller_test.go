package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	auth "gateway/auth"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type mockAuthClient struct {
	resetFn   func(context.Context, *auth.ResetPasswordRequest) (*auth.GenericResponse, error)
	confirmFn func(context.Context, *auth.ConfirmResetPasswordRequest) (*auth.GenericResponse, error)
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

func (m *mockAuthClient) UpdatePassword(context.Context, *auth.UpdatePasswordRequest, ...grpc.CallOption) (*auth.GenericResponse, error) {
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

func resetAuthLimiterStateForTests() {
	resetPasswordIPLimiter = newSlidingWindowLimiter()
	resetPasswordEmailLimiter = newSlidingWindowLimiter()
	confirmResetIPLimiter = newSlidingWindowLimiter()
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
