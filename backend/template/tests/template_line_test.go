package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func TestCreateLine_Success(t *testing.T) {
	srv, mock := setupServer(t)
	r := gin.New()
	srv.SetupRoutes(r)

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM templates`).
		WithArgs("tmpl-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectExec(`INSERT INTO template_lines`).
		WithArgs(sqlmock.AnyArg(), "tmpl-1", "simple", "Prestation", "1", nil, int64(10000), "{}", int32(0), nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"type":"simple","name":"Prestation","quantity":"1","unit_price":10000,"position":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/templates/tmpl-1/lines", strings.NewReader(body))
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

func TestCreateLine_InvalidQuantity(t *testing.T) {
	srv, mock := setupServer(t)
	r := gin.New()
	srv.SetupRoutes(r)

	body := `{"type":"simple","name":"Test","quantity":"not-a-number"}`
	req := httptest.NewRequest(http.MethodPost, "/api/templates/tmpl-1/lines", strings.NewReader(body))
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

func TestDeleteLine_Success(t *testing.T) {
	srv, mock := setupServer(t)
	r := gin.New()
	srv.SetupRoutes(r)

	mock.ExpectExec(`DELETE FROM template_lines`).
		WithArgs("line-1", "tmpl-1", "user-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodDelete, "/api/templates/tmpl-1/lines/line-1", nil)
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
