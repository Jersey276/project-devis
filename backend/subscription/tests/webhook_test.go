package tests

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	stripe "github.com/stripe/stripe-go/v82"
	"github.com/DATA-DOG/go-sqlmock"
	subGrpc "project-devis-subscription/services/grpc"
)

// buildStripeEvent constructs a minimal stripe.Event for testing processWebhookEvent.
func buildStripeEvent(eventType string, rawData []byte) stripe.Event {
	return stripe.Event{
		ID:   "evt_test_123",
		Type: stripe.EventType(eventType),
		Data: &stripe.EventData{Raw: rawData},
	}
}

// ─── CancelSubscription ───────────────────────────────────────────────────────

func TestCancelSubscription_NoSubscription(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT stripe_subscription_id FROM subscriptions WHERE user_id`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"stripe_subscription_id"}))

	resp, err := srv.CancelSubscription(context.Background(), &subGrpc.CancelSubscriptionRequest{UserId: "user-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for user without subscription")
	}
	if resp.Code != 1001 {
		t.Fatalf("expected CodeNotFound (1001), got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCancelSubscription_EmptyStripeId(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT stripe_subscription_id FROM subscriptions WHERE user_id`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"stripe_subscription_id"}).AddRow(""))

	resp, err := srv.CancelSubscription(context.Background(), &subGrpc.CancelSubscriptionRequest{UserId: "user-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for empty stripe_subscription_id")
	}
	if resp.Code != 1001 {
		t.Fatalf("expected CodeNotFound (1001), got %d", resp.Code)
	}
}

// ─── GetAdminStats ────────────────────────────────────────────────────────────

func TestGetAdminStats_ReturnsData(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT p\.tier, COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"tier", "count"}).
			AddRow("free", 5).
			AddRow("pro", 3).
			AddRow("enterprise", 1))

	mock.ExpectQuery(`SELECT COUNT\(\*\), COALESCE`).
		WillReturnRows(sqlmock.NewRows([]string{"count", "revenue"}).AddRow(4, int64(36900)))

	mock.ExpectQuery(`SELECT date_trunc`).
		WillReturnRows(sqlmock.NewRows([]string{"month", "revenue_cents"}).
			AddRow("2026-01-01", int64(8100)).
			AddRow("2026-02-01", int64(9000)))

	resp, err := srv.GetAdminStats(context.Background(), &subGrpc.GetAdminStatsRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if len(resp.PlanDistribution) != 3 {
		t.Errorf("expected 3 plan distribution entries, got %d", len(resp.PlanDistribution))
	}
	if resp.TotalRevenueCents != 36900 {
		t.Errorf("expected total revenue 36900, got %d", resp.TotalRevenueCents)
	}
	if len(resp.MonthlyRevenue) != 2 {
		t.Errorf("expected 2 monthly revenue entries, got %d", len(resp.MonthlyRevenue))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// ─── processWebhookEvent (internal, testable without Stripe SDK) ──────────────

func TestProcessWebhookEvent_DuplicateEvent(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`INSERT INTO subscription_events`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	rawData, _ := json.Marshal(map[string]any{"id": "sub_test", "customer": "cus_test", "status": "active"})
	event := buildStripeEvent("customer.subscription.updated", rawData)

	resp, err := srv.ProcessWebhookEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("duplicate event should return success, got code %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestProcessWebhookEvent_SubscriptionDeleted(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`INSERT INTO subscription_events`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(`UPDATE subscriptions SET status = 'cancelled'`).
		WithArgs("cus_test").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("user-1"))

	rawData, _ := json.Marshal(map[string]any{
		"id":       "sub_test",
		"customer": map[string]any{"id": "cus_test"},
		"status":   "canceled",
	})
	event := buildStripeEvent("customer.subscription.deleted", rawData)

	resp, err := srv.ProcessWebhookEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestProcessWebhookEvent_InvoicePaymentFailed(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`INSERT INTO subscription_events`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`UPDATE subscriptions SET status = 'expired'`).
		WithArgs("cus_test").
		WillReturnResult(sqlmock.NewResult(1, 1))

	rawData, _ := json.Marshal(map[string]any{
		"id":              "in_test",
		"customer":        map[string]any{"id": "cus_test"},
		"period_start":    time.Now().Unix(),
		"period_end":      time.Now().Unix() + 2592000,
	})
	event := buildStripeEvent("invoice.payment_failed", rawData)

	resp, err := srv.ProcessWebhookEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
