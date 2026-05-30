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
	schedule "gateway/schedule"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ─── Mock server ─────────────────────────────────────────────────────────────

type mockScheduleServer struct {
	schedule.UnimplementedScheduleServiceServer
}

func (s *mockScheduleServer) CreateSchedule(_ context.Context, req *schedule.CreateScheduleRequest) (*schedule.CreateScheduleResponse, error) {
	if req.Name == "" {
		return &schedule.CreateScheduleResponse{Success: false, Code: 1003}, nil
	}
	return &schedule.CreateScheduleResponse{Success: true, ScheduleId: "sch-001"}, nil
}

func (s *mockScheduleServer) ListSchedules(_ context.Context, _ *schedule.ListSchedulesRequest) (*schedule.ListSchedulesResponse, error) {
	return &schedule.ListSchedulesResponse{
		Success: true,
		Schedules: []*schedule.ScheduleSummary{
			{ScheduleId: "sch-001", Name: "Éch. 1", Status: "DRAFT", StartMonth: "2026-01", DurationMonths: 3, QuoteId: "q-001"},
		},
	}, nil
}

func (s *mockScheduleServer) GetSchedule(_ context.Context, req *schedule.GetScheduleRequest) (*schedule.GetScheduleResponse, error) {
	if req.ScheduleId == "not-found" {
		return &schedule.GetScheduleResponse{Success: false, Code: 1001}, nil
	}
	return &schedule.GetScheduleResponse{
		Success: true,
		Schedule: &schedule.ScheduleDetails{
			ScheduleId:        req.ScheduleId,
			QuoteId:           "q-001",
			Status:            "DRAFT",
			Name:              "Éch. test",
			StartMonth:        "2026-01",
			DurationMonths:    3,
			Lines:             []*schedule.ScheduleLineSummary{{QuoteLineId: "l-001", PlannedCents: 500, ExpectedCents: 1000}},
			ColumnTotals:      []*schedule.ScheduleColumnTotal{{MonthIndex: 1, AmountCents: 500}},
			QuoteTotalCents:   1000,
			PlannedTotalCents: 500,
		},
	}, nil
}

func (s *mockScheduleServer) UpdateScheduleCell(_ context.Context, req *schedule.UpdateScheduleCellRequest) (*schedule.GenericResponse, error) {
	if req.ScheduleId == "finalized" {
		return &schedule.GenericResponse{Success: false, Code: 1006}, nil
	}
	return &schedule.GenericResponse{Success: true}, nil
}

func (s *mockScheduleServer) ValidateSchedule(_ context.Context, req *schedule.ValidateScheduleRequest) (*schedule.GenericResponse, error) {
	if req.ScheduleId == "unbalanced" {
		return &schedule.GenericResponse{Success: false, Code: 1007}, nil
	}
	return &schedule.GenericResponse{Success: true}, nil
}

// ─── Test helpers ─────────────────────────────────────────────────────────────

func startScheduleTestServer(t *testing.T) (schedule.ScheduleServiceClient, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	schedule.RegisterScheduleServiceServer(s, &mockScheduleServer{})
	go s.Serve(lis)

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		s.Stop()
		t.Fatalf("failed to connect: %v", err)
	}
	return schedule.NewScheduleServiceClient(conn), func() {
		conn.Close()
		s.Stop()
	}
}

// setupScheduleRouter builds a Gin router with a fake user_id middleware,
// matching the production pattern (middleware.CtxUserID = "user_id").
func setupScheduleRouter(client schedule.ScheduleServiceClient) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "user-test")
		c.Next()
	})
	g := r.Group("/schedules")
	g.GET("", func(c *gin.Context) { controllers.ListSchedules(c, client) })
	g.POST("", func(c *gin.Context) { controllers.CreateSchedule(c, client) })
	one := g.Group("/:id")
	one.GET("", func(c *gin.Context) { controllers.GetSchedule(c, client) })
	one.PATCH("/cells", func(c *gin.Context) { controllers.UpdateScheduleCell(c, client) })
	one.POST("/validate", func(c *gin.Context) { controllers.ValidateSchedule(c, client) })
	return r
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestListSchedules_ReturnsSchedules(t *testing.T) {
	client, cleanup := startScheduleTestServer(t)
	defer cleanup()
	r := setupScheduleRouter(client)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/schedules?quote_id=q-001", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	schedules, ok := body["schedules"].([]any)
	if !ok || len(schedules) != 1 {
		t.Fatalf("expected 1 schedule, got %v", body["schedules"])
	}
	s := schedules[0].(map[string]any)
	if s["schedule_id"] != "sch-001" {
		t.Errorf("unexpected schedule_id: %v", s["schedule_id"])
	}
}

func TestCreateSchedule_Success(t *testing.T) {
	client, cleanup := startScheduleTestServer(t)
	defer cleanup()
	r := setupScheduleRouter(client)

	payload := map[string]any{
		"quote_id":        "q-001",
		"name":            "Échéancier principal",
		"start_month":     "2026-01",
		"duration_months": 3,
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/schedules", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["schedule_id"] != "sch-001" {
		t.Errorf("expected schedule_id sch-001, got %v", resp["schedule_id"])
	}
}

func TestCreateSchedule_MissingField_Returns400(t *testing.T) {
	client, cleanup := startScheduleTestServer(t)
	defer cleanup()
	r := setupScheduleRouter(client)

	// name is missing
	payload := map[string]any{"quote_id": "q-001", "start_month": "2026-01", "duration_months": 3}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/schedules", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetSchedule_Success(t *testing.T) {
	client, cleanup := startScheduleTestServer(t)
	defer cleanup()
	r := setupScheduleRouter(client)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/schedules/sch-001", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	sch := resp["schedule"].(map[string]any)
	if sch["status"] != "DRAFT" {
		t.Errorf("expected DRAFT, got %v", sch["status"])
	}
	if sch["quote_total_cents"].(float64) != 1000 {
		t.Errorf("expected quote_total_cents 1000, got %v", sch["quote_total_cents"])
	}
	lines := sch["lines"].([]any)
	if len(lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(lines))
	}
}

func TestGetSchedule_NotFound_Returns404(t *testing.T) {
	client, cleanup := startScheduleTestServer(t)
	defer cleanup()
	r := setupScheduleRouter(client)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/schedules/not-found", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUpdateScheduleCell_Success(t *testing.T) {
	client, cleanup := startScheduleTestServer(t)
	defer cleanup()
	r := setupScheduleRouter(client)

	payload := map[string]any{
		"quote_line_id": "l-001",
		"month_index":   1,
		"amount_eur":    "500.00",
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPatch, "/schedules/sch-001/cells", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestUpdateScheduleCell_Finalized_Returns409(t *testing.T) {
	client, cleanup := startScheduleTestServer(t)
	defer cleanup()
	r := setupScheduleRouter(client)

	payload := map[string]any{
		"quote_line_id": "l-001",
		"month_index":   1,
		"amount_eur":    "100.00",
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPatch, "/schedules/finalized/cells", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestValidateSchedule_Success(t *testing.T) {
	client, cleanup := startScheduleTestServer(t)
	defer cleanup()
	r := setupScheduleRouter(client)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/schedules/sch-001/validate", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestValidateSchedule_Unbalanced_Returns422(t *testing.T) {
	client, cleanup := startScheduleTestServer(t)
	defer cleanup()
	r := setupScheduleRouter(client)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/schedules/unbalanced/validate", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
}
