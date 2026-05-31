package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gateway/controllers"
	export "gateway/export"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type mockExportClient struct {
	quoteResponse    *export.ExportQuoteResponse
	quoteErr         error
	lastQuoteReq     *export.ExportQuoteRequest
	scheduleResponse *export.ExportQuoteResponse
	scheduleErr      error
	lastScheduleReq  *export.ExportScheduleRequest
}

func (m *mockExportClient) ExportQuote(_ context.Context, req *export.ExportQuoteRequest, _ ...grpc.CallOption) (*export.ExportQuoteResponse, error) {
	m.lastQuoteReq = req
	if m.quoteErr != nil {
		return nil, m.quoteErr
	}
	return m.quoteResponse, nil
}

func (m *mockExportClient) ExportSchedule(_ context.Context, req *export.ExportScheduleRequest, _ ...grpc.CallOption) (*export.ExportQuoteResponse, error) {
	m.lastScheduleReq = req
	if m.scheduleErr != nil {
		return nil, m.scheduleErr
	}
	return m.scheduleResponse, nil
}

func setupExportRouter(exportClient export.ExportServiceClient) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "user-test")
		c.Next()
	})
	r.GET("/export/quotes/:id", func(c *gin.Context) {
		controllers.ExportQuote(c, exportClient)
	})
	r.GET("/export/schedules/:id", func(c *gin.Context) {
		controllers.ExportSchedule(c, exportClient)
	})
	return r
}

func TestExportQuote_QuoteRefused(t *testing.T) {
	exportClient := &mockExportClient{
		quoteResponse: &export.ExportQuoteResponse{Success: false, Code: controllers.ExportCodeQuoteRefused},
	}
	r := setupExportRouter(exportClient)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/export/quotes/quote-1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
	if exportClient.lastQuoteReq == nil || exportClient.lastQuoteReq.QuoteId != "quote-1" || exportClient.lastQuoteReq.UserId != "user-test" {
		t.Fatalf("unexpected quote export request: %+v", exportClient.lastQuoteReq)
	}
}

func TestExportSchedule_Success(t *testing.T) {
	exportClient := &mockExportClient{
		scheduleResponse: &export.ExportQuoteResponse{
			Success:  true,
			Code:     0,
			Pdf:      []byte("%PDF-test"),
			Filename: "echeancier-sch-1.pdf",
		},
	}
	r := setupExportRouter(exportClient)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/export/schedules/sch-1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Content-Type"); got != "application/pdf" {
		t.Fatalf("expected application/pdf content type, got %q", got)
	}
	contentDisposition := w.Header().Get("Content-Disposition")
	if !strings.Contains(contentDisposition, "echeancier-sch-1.pdf") {
		t.Fatalf("expected echeancier filename in content disposition, got %q", contentDisposition)
	}
	if exportClient.lastScheduleReq == nil || exportClient.lastScheduleReq.ScheduleId != "sch-1" || exportClient.lastScheduleReq.UserId != "user-test" {
		t.Fatalf("unexpected schedule export request: %+v", exportClient.lastScheduleReq)
	}
}

func TestExportSchedule_NotFound(t *testing.T) {
	exportClient := &mockExportClient{
		scheduleResponse: &export.ExportQuoteResponse{Success: false, Code: controllers.ExportCodeNotFound},
	}
	r := setupExportRouter(exportClient)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/export/schedules/sch-missing", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestExportSchedule_ScheduleUnavailable_Returns502(t *testing.T) {
	exportClient := &mockExportClient{scheduleErr: context.DeadlineExceeded}
	r := setupExportRouter(exportClient)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/export/schedules/sch-1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", w.Code)
	}
}

func TestExportSchedule_ExportUnavailable_Returns502(t *testing.T) {
	exportClient := &mockExportClient{scheduleErr: context.DeadlineExceeded}
	r := setupExportRouter(exportClient)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/export/schedules/sch-1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", w.Code)
	}
}
