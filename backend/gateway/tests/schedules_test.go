package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"gateway/controllers"
	quote "gateway/quote"
	schedule "gateway/schedule"
	gatewaySvc "gateway/services"
	users "gateway/users"

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

// ─── Nop stubs for ancillary clients (used only by the email goroutine) ──────

type nopQuoteClient struct{}

func (nopQuoteClient) CreateQuote(context.Context, *quote.CreateQuoteRequest, ...grpc.CallOption) (*quote.CreateQuoteResponse, error) {
	return nil, nil
}
func (nopQuoteClient) GetQuote(context.Context, *quote.GetQuoteRequest, ...grpc.CallOption) (*quote.GetQuoteResponse, error) {
	return nil, fmt.Errorf("nop")
}
func (nopQuoteClient) ListQuotes(context.Context, *quote.ListQuotesRequest, ...grpc.CallOption) (*quote.ListQuotesResponse, error) {
	return nil, nil
}
func (nopQuoteClient) UpdateQuote(context.Context, *quote.UpdateQuoteRequest, ...grpc.CallOption) (*quote.UpdateQuoteResponse, error) {
	return nil, nil
}
func (nopQuoteClient) DeleteQuote(context.Context, *quote.DeleteQuoteRequest, ...grpc.CallOption) (*quote.GenericResponse, error) {
	return nil, nil
}
func (nopQuoteClient) ArchiveQuote(context.Context, *quote.ArchiveQuoteRequest, ...grpc.CallOption) (*quote.GenericResponse, error) {
	return nil, nil
}
func (nopQuoteClient) RestoreQuote(context.Context, *quote.RestoreQuoteRequest, ...grpc.CallOption) (*quote.GenericResponse, error) {
	return nil, nil
}
func (nopQuoteClient) TrashQuotes(context.Context, *quote.TrashQuotesRequest, ...grpc.CallOption) (*quote.GenericResponse, error) {
	return nil, nil
}
func (nopQuoteClient) DropQuote(context.Context, *quote.DropQuoteRequest, ...grpc.CallOption) (*quote.GenericResponse, error) {
	return nil, nil
}
func (nopQuoteClient) ContinueQuote(context.Context, *quote.ContinueQuoteRequest, ...grpc.CallOption) (*quote.GenericResponse, error) {
	return nil, nil
}
func (nopQuoteClient) ValidateQuote(context.Context, *quote.ValidateQuoteRequest, ...grpc.CallOption) (*quote.GenericResponse, error) {
	return nil, nil
}
func (nopQuoteClient) NegociateQuote(context.Context, *quote.NegociateQuoteRequest, ...grpc.CallOption) (*quote.NegociateQuoteResponse, error) {
	return nil, nil
}
func (nopQuoteClient) CreateQuoteLine(context.Context, *quote.CreateQuoteLineRequest, ...grpc.CallOption) (*quote.CreateQuoteLineResponse, error) {
	return nil, nil
}
func (nopQuoteClient) GetQuoteLine(context.Context, *quote.GetQuoteLineRequest, ...grpc.CallOption) (*quote.GetQuoteLineResponse, error) {
	return nil, nil
}
func (nopQuoteClient) ListQuoteLines(context.Context, *quote.ListQuoteLinesRequest, ...grpc.CallOption) (*quote.ListQuoteLinesResponse, error) {
	return nil, nil
}
func (nopQuoteClient) ListUserQuoteLines(context.Context, *quote.ListUserQuoteLinesRequest, ...grpc.CallOption) (*quote.ListUserQuoteLinesResponse, error) {
	return nil, nil
}
func (nopQuoteClient) UpdateQuoteLine(context.Context, *quote.UpdateQuoteLineRequest, ...grpc.CallOption) (*quote.UpdateQuoteLineResponse, error) {
	return nil, nil
}
func (nopQuoteClient) DeleteQuoteLine(context.Context, *quote.DeleteQuoteLineRequest, ...grpc.CallOption) (*quote.GenericResponse, error) {
	return nil, nil
}
func (nopQuoteClient) CreateFee(context.Context, *quote.CreateFeeRequest, ...grpc.CallOption) (*quote.CreateFeeResponse, error) {
	return nil, nil
}
func (nopQuoteClient) GetFee(context.Context, *quote.GetFeeRequest, ...grpc.CallOption) (*quote.GetFeeResponse, error) {
	return nil, nil
}
func (nopQuoteClient) ListFees(context.Context, *quote.ListFeesRequest, ...grpc.CallOption) (*quote.ListFeesResponse, error) {
	return nil, nil
}
func (nopQuoteClient) UpdateFee(context.Context, *quote.UpdateFeeRequest, ...grpc.CallOption) (*quote.UpdateFeeResponse, error) {
	return nil, nil
}
func (nopQuoteClient) ArchiveFee(context.Context, *quote.ArchiveFeeRequest, ...grpc.CallOption) (*quote.GenericResponse, error) {
	return nil, nil
}

type nopUsersClient struct{}

func (nopUsersClient) CreateUser(context.Context, *users.CreateUserRequest, ...grpc.CallOption) (*users.CreateUserResponse, error) {
	return nil, nil
}
func (nopUsersClient) GetUser(context.Context, *users.GetUserRequest, ...grpc.CallOption) (*users.GetUserResponse, error) {
	return nil, nil
}
func (nopUsersClient) UpdateUser(context.Context, *users.UpdateUserRequest, ...grpc.CallOption) (*users.UpdateUserResponse, error) {
	return nil, nil
}
func (nopUsersClient) DeleteUser(context.Context, *users.DeleteUserRequest, ...grpc.CallOption) (*users.GenericResponse, error) {
	return nil, nil
}
func (nopUsersClient) GetUserAccessInfo(context.Context, *users.GetUserAccessInfoRequest, ...grpc.CallOption) (*users.GetUserAccessInfoResponse, error) {
	return nil, nil
}
func (nopUsersClient) GetUserAccessInfoByEmail(context.Context, *users.GetUserAccessInfoByEmailRequest, ...grpc.CallOption) (*users.GetUserAccessInfoResponse, error) {
	return nil, nil
}
func (nopUsersClient) ListAdminAccounts(context.Context, *users.ListAdminAccountsRequest, ...grpc.CallOption) (*users.ListAdminAccountsResponse, error) {
	return nil, nil
}
func (nopUsersClient) UpdateAdminAccount(context.Context, *users.UpdateAdminAccountRequest, ...grpc.CallOption) (*users.GenericResponse, error) {
	return nil, nil
}
func (nopUsersClient) SuspendAdminAccount(context.Context, *users.SuspendAdminAccountRequest, ...grpc.CallOption) (*users.GenericResponse, error) {
	return nil, nil
}
func (nopUsersClient) TouchUserLastLogin(context.Context, *users.TouchUserLastLoginRequest, ...grpc.CallOption) (*users.GenericResponse, error) {
	return nil, nil
}
func (nopUsersClient) CreateClient(context.Context, *users.CreateClientRequest, ...grpc.CallOption) (*users.CreateClientResponse, error) {
	return nil, nil
}
func (nopUsersClient) GetClient(context.Context, *users.GetClientRequest, ...grpc.CallOption) (*users.GetClientResponse, error) {
	return nil, nil
}
func (nopUsersClient) ListClients(context.Context, *users.ListClientsRequest, ...grpc.CallOption) (*users.ListClientsResponse, error) {
	return nil, nil
}
func (nopUsersClient) UpdateClient(context.Context, *users.UpdateClientRequest, ...grpc.CallOption) (*users.UpdateClientResponse, error) {
	return nil, nil
}
func (nopUsersClient) ArchiveClient(context.Context, *users.ArchiveClientRequest, ...grpc.CallOption) (*users.GenericResponse, error) {
	return nil, nil
}
func (nopUsersClient) CreateAddress(context.Context, *users.CreateAddressRequest, ...grpc.CallOption) (*users.CreateAddressResponse, error) {
	return nil, nil
}
func (nopUsersClient) GetAddress(context.Context, *users.GetAddressRequest, ...grpc.CallOption) (*users.GetAddressResponse, error) {
	return nil, nil
}
func (nopUsersClient) ListAddresses(context.Context, *users.ListAddressesRequest, ...grpc.CallOption) (*users.ListAddressesResponse, error) {
	return nil, nil
}
func (nopUsersClient) UpdateAddress(context.Context, *users.UpdateAddressRequest, ...grpc.CallOption) (*users.UpdateAddressResponse, error) {
	return nil, nil
}
func (nopUsersClient) ArchiveAddress(context.Context, *users.ArchiveAddressRequest, ...grpc.CallOption) (*users.GenericResponse, error) {
	return nil, nil
}
func (nopUsersClient) CreateCountry(context.Context, *users.CreateCountryRequest, ...grpc.CallOption) (*users.CreateCountryResponse, error) {
	return nil, nil
}
func (nopUsersClient) GetCountry(context.Context, *users.GetCountryRequest, ...grpc.CallOption) (*users.GetCountryResponse, error) {
	return nil, nil
}
func (nopUsersClient) ListCountries(context.Context, *users.ListCountriesRequest, ...grpc.CallOption) (*users.ListCountriesResponse, error) {
	return nil, nil
}
func (nopUsersClient) UpdateCountry(context.Context, *users.UpdateCountryRequest, ...grpc.CallOption) (*users.UpdateCountryResponse, error) {
	return nil, nil
}
func (nopUsersClient) DeleteCountry(context.Context, *users.DeleteCountryRequest, ...grpc.CallOption) (*users.GenericResponse, error) {
	return nil, nil
}
func (nopUsersClient) CreateCountryGroup(context.Context, *users.CreateCountryGroupRequest, ...grpc.CallOption) (*users.CreateCountryGroupResponse, error) {
	return nil, nil
}
func (nopUsersClient) GetCountryGroup(context.Context, *users.GetCountryGroupRequest, ...grpc.CallOption) (*users.GetCountryGroupResponse, error) {
	return nil, nil
}
func (nopUsersClient) ListCountryGroups(context.Context, *users.ListCountryGroupsRequest, ...grpc.CallOption) (*users.ListCountryGroupsResponse, error) {
	return nil, nil
}
func (nopUsersClient) UpdateCountryGroup(context.Context, *users.UpdateCountryGroupRequest, ...grpc.CallOption) (*users.UpdateCountryGroupResponse, error) {
	return nil, nil
}
func (nopUsersClient) DeleteCountryGroup(context.Context, *users.DeleteCountryGroupRequest, ...grpc.CallOption) (*users.GenericResponse, error) {
	return nil, nil
}
func (nopUsersClient) AttachCountry(context.Context, *users.AttachCountryRequest, ...grpc.CallOption) (*users.GenericResponse, error) {
	return nil, nil
}
func (nopUsersClient) DetachCountry(context.Context, *users.DetachCountryRequest, ...grpc.CallOption) (*users.GenericResponse, error) {
	return nil, nil
}
func (nopUsersClient) CreateTax(context.Context, *users.CreateTaxRequest, ...grpc.CallOption) (*users.CreateTaxResponse, error) {
	return nil, nil
}
func (nopUsersClient) GetTax(context.Context, *users.GetTaxRequest, ...grpc.CallOption) (*users.GetTaxResponse, error) {
	return nil, nil
}
func (nopUsersClient) ListTaxes(context.Context, *users.ListTaxesRequest, ...grpc.CallOption) (*users.ListTaxesResponse, error) {
	return nil, nil
}
func (nopUsersClient) ListTaxesForUser(context.Context, *users.ListTaxesForUserRequest, ...grpc.CallOption) (*users.ListTaxesResponse, error) {
	return nil, nil
}
func (nopUsersClient) ListTaxesForCountry(context.Context, *users.ListTaxesForCountryRequest, ...grpc.CallOption) (*users.ListTaxesResponse, error) {
	return nil, nil
}
func (nopUsersClient) UpdateTax(context.Context, *users.UpdateTaxRequest, ...grpc.CallOption) (*users.UpdateTaxResponse, error) {
	return nil, nil
}
func (nopUsersClient) DeleteTax(context.Context, *users.DeleteTaxRequest, ...grpc.CallOption) (*users.GenericResponse, error) {
	return nil, nil
}

type nopEmailNotifier struct{}

func (nopEmailNotifier) SendQuoteEmail(context.Context, string, string, string, string, string, []byte) error {
	return nil
}
func (nopEmailNotifier) SendScheduleEmail(context.Context, string, string, string, string, string, string) error {
	return nil
}

var _ quote.QuoteServiceClient = nopQuoteClient{}
var _ users.UserServiceClient = nopUsersClient{}
var _ gatewaySvc.EmailNotifier = nopEmailNotifier{}

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
	g.GET("", func(c *gin.Context) { controllers.ListSchedules(c, client, nopQuoteClient{}) })
	g.POST("", func(c *gin.Context) { controllers.CreateSchedule(c, client, nopQuoteClient{}) })
	one := g.Group("/:id")
	one.GET("", func(c *gin.Context) { controllers.GetSchedule(c, client, nopQuoteClient{}) })
	one.PATCH("/cells", func(c *gin.Context) { controllers.UpdateScheduleCell(c, client) })
	one.POST("/validate", func(c *gin.Context) {
		controllers.ValidateSchedule(c, client, nopQuoteClient{}, nopUsersClient{}, nopEmailNotifier{})
	})
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
