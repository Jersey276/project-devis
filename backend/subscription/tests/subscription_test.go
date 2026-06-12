package tests

import (
	"context"
	"testing"

	"project-devis-subscription/actions"
	subGrpc "project-devis-subscription/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

// ─── Helpers ─────────────────────────────────────────────────────────────────

func setupServer(t *testing.T) (*actions.Server, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return actions.NewServer(db, "test-webhook-secret"), mock
}

// ─── ListPlans ────────────────────────────────────────────────────────────────

func TestListPlans_ReturnsPlans(t *testing.T) {
	srv, mock := setupServer(t)

	rows := sqlmock.NewRows([]string{"plan_id", "name", "tier", "price_cents", "billing_cycle", "features", "active", "stripe_price_id"}).
		AddRow(1, "Free", "free", 0, "none", `{"max_schedules":3}`, true, "").
		AddRow(2, "Pro", "pro", 900, "monthly", `{"max_schedules":-1}`, true, "").
		AddRow(3, "Enterprise", "enterprise", 4900, "monthly", `{"max_schedules":-1}`, true, "")

	mock.ExpectQuery(`SELECT plan_id, name, tier, price_cents, billing_cycle, features::text, active, COALESCE`).
		WillReturnRows(rows)

	resp, err := srv.ListPlans(context.Background(), &subGrpc.ListPlansRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if len(resp.Plans) != 3 {
		t.Fatalf("expected 3 plans, got %d", len(resp.Plans))
	}
	if resp.Plans[0].Tier != "free" {
		t.Errorf("expected first plan tier 'free', got %s", resp.Plans[0].Tier)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// ─── GetUserSubscription ─────────────────────────────────────────────────────

func TestGetUserSubscription_ExistingUser(t *testing.T) {
	srv, mock := setupServer(t)

	rows := sqlmock.NewRows([]string{
		"subscription_id", "user_id", "plan_id", "tier", "status",
		"current_period_start", "current_period_end", "created_at", "updated_at",
	}).AddRow("sub-1", "user-1", 2, "pro", "active", "2026-01-01", nil, "2026-01-01", "2026-01-01")

	mock.ExpectQuery(`SELECT s.subscription_id`).
		WithArgs("user-1").
		WillReturnRows(rows)

	resp, err := srv.GetUserSubscription(context.Background(), &subGrpc.GetUserSubscriptionRequest{UserId: "user-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Subscription.Tier != "pro" {
		t.Errorf("expected tier 'pro', got %s", resp.Subscription.Tier)
	}
}

func TestGetUserSubscription_NoRow_ReturnsSyntheticFree(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT s.subscription_id`).
		WithArgs("new-user").
		WillReturnRows(sqlmock.NewRows([]string{}))

	resp, err := srv.GetUserSubscription(context.Background(), &subGrpc.GetUserSubscriptionRequest{UserId: "new-user"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success even for missing row, got code %d", resp.Code)
	}
	if resp.Subscription == nil {
		t.Fatal("expected synthetic subscription, got nil")
	}
	if resp.Subscription.Tier != "free" {
		t.Errorf("expected synthetic tier 'free', got %s", resp.Subscription.Tier)
	}
	if resp.Subscription.UserId != "new-user" {
		t.Errorf("expected user_id 'new-user', got %s", resp.Subscription.UserId)
	}
}

// ─── AssignPlan ───────────────────────────────────────────────────────────────

func TestAssignPlan_NewSubscription(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT tier FROM plans WHERE plan_id = \$1 AND active = TRUE`).
		WithArgs(int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"tier"}).AddRow("pro"))

	mock.ExpectExec(`INSERT INTO subscriptions`).
		WithArgs(sqlmock.AnyArg(), "user-1", int32(2)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.AssignPlan(context.Background(), &subGrpc.AssignPlanRequest{UserId: "user-1", PlanId: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.NewTier != "pro" {
		t.Errorf("expected new_tier 'pro', got %s", resp.NewTier)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAssignPlan_PlanNotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT tier FROM plans WHERE plan_id = \$1 AND active = TRUE`).
		WithArgs(int32(99)).
		WillReturnRows(sqlmock.NewRows([]string{"tier"}))

	resp, err := srv.AssignPlan(context.Background(), &subGrpc.AssignPlanRequest{UserId: "user-1", PlanId: 99})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for nonexistent plan")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound (%d), got %d", actions.CodeNotFound, resp.Code)
	}
}

func TestAssignPlan_InvalidInput_MissingUserId(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.AssignPlan(context.Background(), &subGrpc.AssignPlanRequest{UserId: "", PlanId: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for empty user_id")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput (%d), got %d", actions.CodeInvalidInput, resp.Code)
	}
}

func TestAssignPlan_InvalidInput_MissingPlanId(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.AssignPlan(context.Background(), &subGrpc.AssignPlanRequest{UserId: "user-1", PlanId: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for plan_id = 0")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput (%d), got %d", actions.CodeInvalidInput, resp.Code)
	}
}
