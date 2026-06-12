package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCreateTemplate_Success(t *testing.T) {
	srv, mock := setupServer(t)
	r := gin.New()
	srv.SetupRoutes(r)

	mock.ExpectExec(`INSERT INTO templates`).
		WithArgs(sqlmock.AnyArg(), "user-1", "quote_document", "quote", "Mon Template").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"template_type":"quote_document","target_resource":"quote","name":"Mon Template"}`
	req := httptest.NewRequest(http.MethodPost, "/api/templates", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "user-1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateTemplate_InvalidType(t *testing.T) {
	srv, mock := setupServer(t)
	r := gin.New()
	srv.SetupRoutes(r)

	body := `{"template_type":"invalid_type","target_resource":"quote","name":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/templates", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "user-1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestCreateTemplate_MissingUserID(t *testing.T) {
	srv, mock := setupServer(t)
	r := gin.New()
	srv.SetupRoutes(r)

	body := `{"template_type":"quote_document","target_resource":"quote","name":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/templates", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestGetTemplate_NotFound(t *testing.T) {
	srv, mock := setupServer(t)
	r := gin.New()
	srv.SetupRoutes(r)

	mock.ExpectQuery(`SELECT template_id`).
		WithArgs("unknown-id", "user-1").
		WillReturnRows(sqlmock.NewRows(nil))

	req := httptest.NewRequest(http.MethodGet, "/api/templates/unknown-id", nil)
	req.Header.Set("X-User-Id", "user-1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteTemplate_Success(t *testing.T) {
	srv, mock := setupServer(t)
	r := gin.New()
	srv.SetupRoutes(r)

	mock.ExpectExec(`DELETE FROM templates`).
		WithArgs("tmpl-1", "user-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodDelete, "/api/templates/tmpl-1", nil)
	req.Header.Set("X-User-Id", "user-1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestArchiveTemplate_Success(t *testing.T) {
	srv, mock := setupServer(t)
	r := gin.New()
	srv.SetupRoutes(r)

	mock.ExpectExec(`UPDATE templates SET archived_at`).
		WithArgs("tmpl-1", "user-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodPost, "/api/templates/tmpl-1/archive", nil)
	req.Header.Set("X-User-Id", "user-1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
