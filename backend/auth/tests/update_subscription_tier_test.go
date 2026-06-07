package tests

import (
	"context"
	"testing"

	"project-devis-auth/actions"
	authGrpc "project-devis-auth/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUpdateSubscriptionTier_Success(t *testing.T) {
	srv, mock := setupServer(t, &MockUserClient{})

	mock.ExpectExec(`UPDATE auth SET subscription_tier`).
		WithArgs("pro", "user-123").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.UpdateSubscriptionTier(context.Background(), &authGrpc.UpdateSubscriptionTierRequest{
		UserId: "user-123",
		Tier:   "pro",
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

func TestUpdateSubscriptionTier_Enterprise(t *testing.T) {
	srv, mock := setupServer(t, &MockUserClient{})

	mock.ExpectExec(`UPDATE auth SET subscription_tier`).
		WithArgs("enterprise", "user-456").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.UpdateSubscriptionTier(context.Background(), &authGrpc.UpdateSubscriptionTierRequest{
		UserId: "user-456",
		Tier:   "enterprise",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

func TestUpdateSubscriptionTier_InvalidTier(t *testing.T) {
	srv, _ := setupServer(t, &MockUserClient{})

	resp, err := srv.UpdateSubscriptionTier(context.Background(), &authGrpc.UpdateSubscriptionTierRequest{
		UserId: "user-123",
		Tier:   "premium",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for invalid tier")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput (%d), got %d", actions.CodeInvalidInput, resp.Code)
	}
}

func TestUpdateSubscriptionTier_UserNotFound(t *testing.T) {
	srv, mock := setupServer(t, &MockUserClient{})

	mock.ExpectExec(`UPDATE auth SET subscription_tier`).
		WithArgs("pro", "nonexistent-user").
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, err := srv.UpdateSubscriptionTier(context.Background(), &authGrpc.UpdateSubscriptionTierRequest{
		UserId: "nonexistent-user",
		Tier:   "pro",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for nonexistent user")
	}
	if resp.Code != actions.CodeUserNotFound {
		t.Fatalf("expected CodeUserNotFound (%d), got %d", actions.CodeUserNotFound, resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateSubscriptionTier_BackToFree(t *testing.T) {
	srv, mock := setupServer(t, &MockUserClient{})

	mock.ExpectExec(`UPDATE auth SET subscription_tier`).
		WithArgs("free", "user-123").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.UpdateSubscriptionTier(context.Background(), &authGrpc.UpdateSubscriptionTierRequest{
		UserId: "user-123",
		Tier:   "free",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}
