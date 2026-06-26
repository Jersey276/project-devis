package iopole

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"project-devis-invoice/pdp"
)

// stubDocs is a DocumentSource returning fixed bytes (no export call).
type stubDocs struct {
	pdf []byte
	err error
}

func (s stubDocs) FetchFacturX(context.Context, string, string) ([]byte, error) {
	return s.pdf, s.err
}

// newTestClient builds a Client pointed at a test server with a static token,
// bypassing the real OAuth provider.
func newTestClient(baseURL string, docs DocumentSource) *Client {
	return &Client{
		http:       http.DefaultClient,
		baseURL:    baseURL,
		customerID: "cust-1",
		token:      staticTokenProvider("test-token"),
		docs:       docs,
	}
}

func newTestDirectory(baseURL string) *Directory {
	return &Directory{
		http:       http.DefaultClient,
		baseURL:    baseURL,
		customerID: "cust-1",
		token:      staticTokenProvider("test-token"),
	}
}

func TestSubmit_DepositsFacturXAndReturnsID(t *testing.T) {
	var gotAuth, gotCustomer, gotContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/invoice" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		gotCustomer = r.Header.Get("customer-id")
		gotContentType = r.Header.Get("Content-Type")
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		if _, _, err := r.FormFile("file"); err != nil {
			t.Errorf("missing file part: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"type":"INVOICE","id":"abc-123"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, stubDocs{pdf: []byte("%PDF-1.4 fake")})
	res, err := c.Submit(context.Background(), pdp.SubmitInput{
		InvoiceID: "inv1", UserID: "u1", InvoiceNumber: "2026-0001",
	})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if res.SubmissionID != "abc-123" {
		t.Errorf("SubmissionID = %q, want abc-123", res.SubmissionID)
	}
	if res.Status != pdp.PlatformSubmitted {
		t.Errorf("Status = %q, want SUBMITTED", res.Status)
	}
	if gotAuth != "Bearer test-token" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if gotCustomer != "cust-1" {
		t.Errorf("customer-id = %q", gotCustomer)
	}
	if !strings.HasPrefix(gotContentType, "multipart/form-data") {
		t.Errorf("Content-Type = %q", gotContentType)
	}
}

func TestSubmit_Non201IsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		_, _ = w.Write([]byte("too large"))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, stubDocs{pdf: []byte("x")})
	if _, err := c.Submit(context.Background(), pdp.SubmitInput{InvoiceNumber: "n"}); err == nil {
		t.Fatal("expected error on 413, got nil")
	}
}

func TestSubmit_DocumentErrorAbortsBeforeNetwork(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	defer srv.Close()

	c := newTestClient(srv.URL, stubDocs{err: context.DeadlineExceeded})
	if _, err := c.Submit(context.Background(), pdp.SubmitInput{}); err == nil {
		t.Fatal("expected error when document fetch fails")
	}
	if called {
		t.Error("platform must not be called when the document cannot be fetched")
	}
}

func TestFetchStatus_PicksMostRecentAndMaps(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/invoice/sub-1/status-history" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		// Out-of-order on purpose: newest (APPROVED) is not last.
		_, _ = w.Write([]byte(`[
			{"date":"2025-01-01T10:00:00Z","status":{"code":"SUBMITTED"}},
			{"date":"2025-01-03T10:00:00Z","status":{"code":"APPROVED"}},
			{"date":"2025-01-02T10:00:00Z","status":{"code":"RECEIVED"}}
		]`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, stubDocs{})
	got, err := c.FetchStatus(context.Background(), "sub-1")
	if err != nil {
		t.Fatalf("FetchStatus: %v", err)
	}
	if got != pdp.PlatformApproved {
		t.Errorf("got %q, want APPROVED", got)
	}
}

func TestFetchStatus_EmptyHistoryIsUnknown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, stubDocs{})
	got, err := c.FetchStatus(context.Background(), "sub-1")
	if err != nil {
		t.Fatalf("FetchStatus: %v", err)
	}
	if got != pdp.PlatformUnknown {
		t.Errorf("got %q, want UNKNOWN", got)
	}
}

func TestResolve_FoundAndNotFound(t *testing.T) {
	// Iopole rejects 14-digit SIRETs; Resolve strips the NIC suffix and queries by SIREN (9 digits).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") == "111111111" {
			_, _ = w.Write([]byte(`{"data":[{"businessEntityId":"be-1","name":"ACME"}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	d := newTestDirectory(srv.URL)

	got, err := d.Resolve(context.Background(), "11111111100014")
	if err != nil {
		t.Fatalf("Resolve(found): %v", err)
	}
	if got.RoutingID != "be-1" || got.PlatformName != "ACME" {
		t.Errorf("routing = %+v", got)
	}

	if _, err := d.Resolve(context.Background(), "99999999900014"); err != pdp.ErrRecipientNotFound {
		t.Errorf("Resolve(absent) err = %v, want ErrRecipientNotFound", err)
	}
}

func TestMapStatus_AllIopoleCodes(t *testing.T) {
	cases := map[string]pdp.PlatformStatus{
		"SUBMITTED":          pdp.PlatformSubmitted,
		"ISSUED":             pdp.PlatformSubmitted,
		"RECEIVED":           pdp.PlatformReceived,
		"MADE_AVAILABLE":     pdp.PlatformReceived,
		"IN_HAND":            pdp.PlatformReceived,
		"APPROVED":           pdp.PlatformApproved,
		"PARTIALLY_APPROVED": pdp.PlatformApproved,
		"COMPLETED":          pdp.PlatformApproved,
		"PAYMENT_SENT":       pdp.PlatformCollected,
		"PAYMENT_RECEIVED":   pdp.PlatformCollected,
		"REJECTED":           pdp.PlatformRejected,
		"REFUSED":            pdp.PlatformRejected,
		"UNACCEPTABLE":       pdp.PlatformRejected,
		"DISPUTED":           pdp.PlatformUnknown,
		"SUSPENDED":          pdp.PlatformUnknown,
		"WAT_IS_THIS":        pdp.PlatformUnknown,
		"":                   pdp.PlatformUnknown,
	}
	for code, want := range cases {
		if got := mapStatus(code); got != want {
			t.Errorf("mapStatus(%q) = %q, want %q", code, got, want)
		}
	}
}
