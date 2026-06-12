package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"gateway/controllers"
	sub "gateway/subscription"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ─── Mock server ─────────────────────────────────────────────────────────────

type mockSubscriptionServer struct {
	sub.UnimplementedSubscriptionServiceServer
}

func (s *mockSubscriptionServer) ListPlans(_ context.Context, _ *sub.ListPlansRequest) (*sub.ListPlansResponse, error) {
	return &sub.ListPlansResponse{
		Success: true,
		Plans: []*sub.Plan{
			{PlanId: 1, Name: "Free", Tier: "free", PriceCents: 0, BillingCycle: "none", Features: `{"max_schedules":3}`},
			{PlanId: 2, Name: "Pro", Tier: "pro", PriceCents: 900, BillingCycle: "monthly", Features: `{"max_schedules":-1}`},
		},
	}, nil
}

func (s *mockSubscriptionServer) GetUserSubscription(_ context.Context, req *sub.GetUserSubscriptionRequest) (*sub.GetUserSubscriptionResponse, error) {
	return &sub.GetUserSubscriptionResponse{
		Success: true,
		Subscription: &sub.Subscription{
			SubscriptionId: "sub-001",
			UserId:         req.UserId,
			Tier:           "pro",
			Status:         "active",
		},
	}, nil
}

func (s *mockSubscriptionServer) ListSubscriptions(_ context.Context, _ *sub.ListSubscriptionsRequest) (*sub.ListSubscriptionsResponse, error) {
	return &sub.ListSubscriptionsResponse{
		Success: true,
		Subscriptions: []*sub.Subscription{
			{SubscriptionId: "sub-001", UserId: "user-1", Tier: "pro", Status: "active"},
		},
		Total: 1,
	}, nil
}

func (s *mockSubscriptionServer) AssignPlan(_ context.Context, req *sub.AssignPlanRequest) (*sub.AssignPlanResponse, error) {
	if req.PlanId == 99 {
		return &sub.AssignPlanResponse{Success: false, Code: 1001}, nil
	}
	return &sub.AssignPlanResponse{Success: true, NewTier: "pro"}, nil
}

func (s *mockSubscriptionServer) CreatePaymentIntent(_ context.Context, _ *sub.CreatePaymentIntentRequest) (*sub.CreatePaymentIntentResponse, error) {
	return &sub.CreatePaymentIntentResponse{
		Success:              true,
		ClientSecret:         "pi_test_secret_xxx",
		StripeSubscriptionId: "sub_test_123",
	}, nil
}

func (s *mockSubscriptionServer) HandleStripeWebhook(_ context.Context, _ *sub.HandleStripeWebhookRequest) (*sub.GenericResponse, error) {
	return &sub.GenericResponse{Success: true}, nil
}

func (s *mockSubscriptionServer) CancelSubscription(_ context.Context, _ *sub.CancelSubscriptionRequest) (*sub.GenericResponse, error) {
	return &sub.GenericResponse{Success: true}, nil
}

func (s *mockSubscriptionServer) GetAdminStats(_ context.Context, _ *sub.GetAdminStatsRequest) (*sub.AdminStatsResponse, error) {
	return &sub.AdminStatsResponse{
		Success:                  true,
		TotalActiveSubscriptions: 5,
		TotalRevenueCents:        45000,
		PlanDistribution: []*sub.PlanDistributionEntry{
			{Tier: "free", Count: 3},
			{Tier: "pro", Count: 2},
		},
		MonthlyRevenue: []*sub.MonthlyRevenueEntry{
			{Month: "2026-01", RevenueCents: 1800},
		},
	}, nil
}

// ─── Test helpers ─────────────────────────────────────────────────────────────

func startSubscriptionTestServer(t *testing.T) (sub.SubscriptionServiceClient, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	sub.RegisterSubscriptionServiceServer(s, &mockSubscriptionServer{})
	go s.Serve(lis)

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		s.Stop()
		t.Fatalf("failed to connect: %v", err)
	}
	return sub.NewSubscriptionServiceClient(conn), func() {
		conn.Close()
		s.Stop()
	}
}

func setupSubscriptionRouter(client sub.SubscriptionServiceClient) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Webhook route — no auth middleware
	webhooks := r.Group("/webhooks")
	webhooks.POST("/stripe", func(c *gin.Context) { controllers.HandleStripeWebhook(c, client) })

	// Authenticated routes
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "user-test")
		c.Set("email", "test@example.com")
		c.Next()
	})

	plans := r.Group("/plans")
	plans.GET("", func(c *gin.Context) { controllers.ListPlans(c, client) })

	subscriptions := r.Group("/subscriptions")
	subscriptions.GET("/me", func(c *gin.Context) { controllers.GetMySubscription(c, client) })
	subscriptions.POST("/payment-intent", func(c *gin.Context) { controllers.CreatePaymentIntent(c, client) })
	subscriptions.POST("/cancel", func(c *gin.Context) { controllers.CancelSubscription(c, client) })
	subscriptions.GET("/admin", func(c *gin.Context) { controllers.ListSubscriptionsAdmin(c, client) })
	subscriptions.GET("/admin/stats", func(c *gin.Context) { controllers.GetAdminStats(c, client) })
	subscriptions.POST("/admin/:userId/plan", func(c *gin.Context) { controllers.AssignPlan(c, client) })

	return r
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestListPlans_Returns200WithPlans(t *testing.T) {
	client, cleanup := startSubscriptionTestServer(t)
	defer cleanup()
	r := setupSubscriptionRouter(client)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/plans", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	plans, ok := body["plans"].([]any)
	if !ok || len(plans) != 2 {
		t.Fatalf("expected 2 plans, got %v", body["plans"])
	}
}

func TestGetMySubscription_Returns200(t *testing.T) {
	client, cleanup := startSubscriptionTestServer(t)
	defer cleanup()
	r := setupSubscriptionRouter(client)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/subscriptions/me", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	s, ok := body["subscription"].(map[string]any)
	if !ok {
		t.Fatalf("expected subscription object, got %v", body["subscription"])
	}
	if s["tier"] != "pro" {
		t.Errorf("expected tier 'pro', got %v", s["tier"])
	}
}

func TestListSubscriptionsAdmin_Returns200(t *testing.T) {
	client, cleanup := startSubscriptionTestServer(t)
	defer cleanup()
	r := setupSubscriptionRouter(client)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/subscriptions/admin", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	subs, ok := body["subscriptions"].([]any)
	if !ok || len(subs) != 1 {
		t.Fatalf("expected 1 subscription, got %v", body["subscriptions"])
	}
}

func TestAssignPlan_Success(t *testing.T) {
	client, cleanup := startSubscriptionTestServer(t)
	defer cleanup()
	r := setupSubscriptionRouter(client)

	payload := map[string]any{"plan_id": 2}
	b, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/subscriptions/admin/user-1/plan", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}
}

func TestAssignPlan_PlanNotFound_Returns404(t *testing.T) {
	client, cleanup := startSubscriptionTestServer(t)
	defer cleanup()
	r := setupSubscriptionRouter(client)

	payload := map[string]any{"plan_id": 99}
	b, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/subscriptions/admin/user-1/plan", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestAssignPlan_InvalidBody_Returns400(t *testing.T) {
	client, cleanup := startSubscriptionTestServer(t)
	defer cleanup()
	r := setupSubscriptionRouter(client)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/subscriptions/admin/user-1/plan", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestCreatePaymentIntent_Returns200(t *testing.T) {
	client, cleanup := startSubscriptionTestServer(t)
	defer cleanup()
	r := setupSubscriptionRouter(client)

	payload := map[string]any{"plan_id": 2}
	b, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/subscriptions/payment-intent", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["client_secret"] == "" || body["client_secret"] == nil {
		t.Errorf("expected client_secret in response, got %v", body)
	}
}

func TestCreatePaymentIntent_MissingBody_Returns400(t *testing.T) {
	client, cleanup := startSubscriptionTestServer(t)
	defer cleanup()
	r := setupSubscriptionRouter(client)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/subscriptions/payment-intent", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestHandleStripeWebhook_Returns200(t *testing.T) {
	client, cleanup := startSubscriptionTestServer(t)
	defer cleanup()
	r := setupSubscriptionRouter(client)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/webhooks/stripe", bytes.NewReader([]byte(`{"type":"test"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Stripe-Signature", "t=123,v1=abc")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestCancelSubscription_Returns200(t *testing.T) {
	client, cleanup := startSubscriptionTestServer(t)
	defer cleanup()
	r := setupSubscriptionRouter(client)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/subscriptions/cancel", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestGetAdminStats_Returns200(t *testing.T) {
	client, cleanup := startSubscriptionTestServer(t)
	defer cleanup()
	r := setupSubscriptionRouter(client)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/subscriptions/admin/stats", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["total_active_subscriptions"] == nil {
		t.Errorf("expected total_active_subscriptions in response")
	}
}
