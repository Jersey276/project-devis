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

	rows := sqlmock.NewRows([]string{"plan_id", "name", "tier", "price_cents", "billing_cycle", "features", "active", "stripe_price_id", "stripe_product_id"}).
		AddRow(1, "Free", "free", 0, "none", `{"max_schedules":3}`, true, "", "").
		AddRow(2, "Pro", "pro", 900, "monthly", `{"max_schedules":-1}`, true, "", "").
		AddRow(3, "Enterprise", "enterprise", 4900, "monthly", `{"max_schedules":-1}`, true, "", "")

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

// ─── UpdatePlan ───────────────────────────────────────────────────────────────

// Without a Stripe key, UpdatePlan must still update the DB and return the plan.
func TestUpdatePlan_DBOnly_NoStripeKey(t *testing.T) {
	srv, mock := setupServer(t)

	// SELECT current plan state
	mock.ExpectQuery(`SELECT name, price_cents, billing_cycle, stripe_price_id, stripe_product_id FROM plans WHERE plan_id`).
		WithArgs(int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"name", "price_cents", "billing_cycle", "stripe_price_id", "stripe_product_id"}).
			AddRow("Pro", int32(900), "monthly", nil, nil))

	// UPDATE
	mock.ExpectExec(`UPDATE plans`).
		WithArgs("Pro Updated", int32(1200), "monthly", sqlmock.AnyArg(), sqlmock.AnyArg(), `{"max_schedules":-1}`, int32(2)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// SELECT read-back
	mock.ExpectQuery(`SELECT plan_id, name, tier, price_cents, billing_cycle, features::text, active`).
		WithArgs(int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"plan_id", "name", "tier", "price_cents", "billing_cycle", "features", "active", "stripe_price_id", "stripe_product_id"}).
			AddRow(2, "Pro Updated", "pro", int32(1200), "monthly", `{"max_schedules":-1}`, true, "", ""))

	resp, err := srv.UpdatePlan(context.Background(), &subGrpc.UpdatePlanRequest{
		PlanId:       2,
		Name:         "Pro Updated",
		PriceCents:   1200,
		BillingCycle: "monthly",
		Features:     `{"max_schedules":-1}`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Plan.Name != "Pro Updated" {
		t.Errorf("expected name 'Pro Updated', got %s", resp.Plan.Name)
	}
	if resp.Plan.PriceCents != 1200 {
		t.Errorf("expected price_cents 1200, got %d", resp.Plan.PriceCents)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdatePlan_InvalidInput_MissingPlanId(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.UpdatePlan(context.Background(), &subGrpc.UpdatePlanRequest{PlanId: 0, Name: "x"})
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

func TestUpdatePlan_InvalidInput_MissingName(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.UpdatePlan(context.Background(), &subGrpc.UpdatePlanRequest{PlanId: 2, Name: ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for empty name")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput (%d), got %d", actions.CodeInvalidInput, resp.Code)
	}
}

func TestUpdatePlan_InvalidInput_BadJSON(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.UpdatePlan(context.Background(), &subGrpc.UpdatePlanRequest{PlanId: 2, Name: "Pro", Features: "not-json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for invalid JSON features")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput (%d), got %d", actions.CodeInvalidInput, resp.Code)
	}
}

// ─── ChangePlan ───────────────────────────────────────────────────────────────

func TestChangePlan_InvalidInput_MissingUserId(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.ChangePlan(context.Background(), &subGrpc.ChangePlanRequest{UserId: "", PlanId: 3})
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

func TestChangePlan_InvalidInput_MissingPlanId(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.ChangePlan(context.Background(), &subGrpc.ChangePlanRequest{UserId: "user-1", PlanId: 0})
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

// ─── ReactivateSubscription ───────────────────────────────────────────────────

func TestReactivateSubscription_InvalidInput_MissingUserId(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.ReactivateSubscription(context.Background(), &subGrpc.ReactivateSubscriptionRequest{UserId: ""})
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

func TestReactivateSubscription_NotCancelling(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT stripe_subscription_id, cancel_at_period_end FROM subscriptions WHERE user_id`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"stripe_subscription_id", "cancel_at_period_end"}).
			AddRow("sub_stripe_123", false))

	resp, err := srv.ReactivateSubscription(context.Background(), &subGrpc.ReactivateSubscriptionRequest{UserId: "user-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure when cancel_at_period_end is false")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput (%d), got %d", actions.CodeInvalidInput, resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
