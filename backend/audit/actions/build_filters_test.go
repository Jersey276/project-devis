package actions

import (
	"fmt"
	"strings"
	"testing"

	auditGrpc "project-devis-audit/services/grpc"
)

// shared sentinel error for in-package tests
var errDBDown = fmt.Errorf("db connection refused")

func TestBuildFilters_Nil(t *testing.T) {
	where, args := buildFilters(nil)
	if where != "" {
		t.Fatalf("expected empty clause, got %q", where)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
}

func TestBuildFilters_UserID(t *testing.T) {
	where, args := buildFilters(&auditGrpc.ActivityLogFilters{UserId: "u1"})
	if !strings.Contains(where, "user_id") {
		t.Fatalf("expected user_id clause, got %q", where)
	}
	if len(args) != 1 || args[0] != "u1" {
		t.Fatalf("expected args=[u1], got %v", args)
	}
}

func TestBuildFilters_URLContains(t *testing.T) {
	where, args := buildFilters(&auditGrpc.ActivityLogFilters{UrlContains: "quotes"})
	if !strings.Contains(where, "ILIKE") {
		t.Fatalf("expected ILIKE clause, got %q", where)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %v", args)
	}
	if !strings.Contains(fmt.Sprintf("%v", args[0]), "quotes") {
		t.Fatalf("expected arg to contain 'quotes', got %v", args[0])
	}
}

func TestBuildFilters_UserIDAndURL(t *testing.T) {
	where, args := buildFilters(&auditGrpc.ActivityLogFilters{UserId: "u1", UrlContains: "quotes"})
	if !strings.Contains(where, "OR") {
		t.Fatalf("expected OR clause, got %q", where)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %v", args)
	}
}

func TestBuildFilters_RespStatuses(t *testing.T) {
	where, args := buildFilters(&auditGrpc.ActivityLogFilters{RespStatuses: []int32{500, 502}})
	if !strings.Contains(where, "IN") {
		t.Fatalf("expected IN clause, got %q", where)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %v", args)
	}
}

func TestBuildFilters_DateRange(t *testing.T) {
	where, args := buildFilters(&auditGrpc.ActivityLogFilters{
		DateFrom: "2024-01-01",
		DateTo:   "2024-12-31",
	})
	if !strings.Contains(where, ">=") || !strings.Contains(where, "<=") {
		t.Fatalf("expected date range clauses, got %q", where)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %v", args)
	}
}

func TestBuildFilters_AllFilters(t *testing.T) {
	where, args := buildFilters(&auditGrpc.ActivityLogFilters{
		UserId:       "u1",
		RespStatuses: []int32{200},
		DateFrom:     "2024-01-01",
		DateTo:       "2024-12-31",
	})
	if !strings.HasPrefix(where, " WHERE ") {
		t.Fatalf("expected WHERE clause, got %q", where)
	}
	if len(args) < 4 {
		t.Fatalf("expected at least 4 args, got %v", args)
	}
}
