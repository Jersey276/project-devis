package iopole

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"project-devis-invoice/pdp"
)

func newTestReporter(baseURL string) *Reporter {
	return &Reporter{
		http:       http.DefaultClient,
		baseURL:    baseURL,
		customerID: "cust-1",
		token:      staticTokenProvider("test-token"),
	}
}

func TestReporter_Submit_OK(t *testing.T) {
	var gotAuth, gotCustomer, gotPath string
	var gotBody transactionBody

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotCustomer = r.Header.Get("customer-id")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	r := newTestReporter(srv.URL)
	res, err := r.SubmitReport(context.Background(), pdp.SubmitReportInput{
		UserID:      "u1",
		IssuerSIREN: "123456789",
		Kind:        pdp.ReportTransaction,
		Period:      pdp.ReportPeriod{Year: 2026, Month: 3},
		Lines: []pdp.ReportLine{
			{TaxRate: "20", CountryCode: "FR", BaseHTCents: 100000, VATCents: 20000},
		},
		TotalHTCents:  100000,
		TotalVATCents: 20000,
	})
	if err != nil {
		t.Fatalf("SubmitReport: %v", err)
	}
	if res.Status != pdp.PlatformSubmitted {
		t.Errorf("Status = %q, want SUBMITTED", res.Status)
	}
	if res.ReportID != "" {
		t.Errorf("ReportID = %q, want empty", res.ReportID)
	}

	if gotPath != "/v1/reporting/transaction/scheme/0002/value/123456789" {
		t.Errorf("path = %q", gotPath)
	}
	if gotAuth != "Bearer test-token" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if gotCustomer != "cust-1" {
		t.Errorf("customer-id = %q", gotCustomer)
	}
	if gotBody.TransactionDate != "2026-03-31" {
		t.Errorf("transactionDate = %q, want 2026-03-31", gotBody.TransactionDate)
	}
	if len(gotBody.Transactions) != 1 {
		t.Fatalf("transactions count = %d, want 1", len(gotBody.Transactions))
	}
	tx := gotBody.Transactions[0]
	if tx.CategoryCode != "TPS1" {
		t.Errorf("categoryCode = %q, want TPS1", tx.CategoryCode)
	}
	if tx.Currency != "EUR" {
		t.Errorf("currency = %q, want EUR", tx.Currency)
	}
	if tx.Monetary.TaxBasisTotalAmount.Amount != 1000.0 {
		t.Errorf("taxBasisTotalAmount = %v, want 1000.0", tx.Monetary.TaxBasisTotalAmount.Amount)
	}
	if tx.Monetary.TaxTotalAmount.Amount != 200.0 {
		t.Errorf("taxTotalAmount = %v, want 200.0", tx.Monetary.TaxTotalAmount.Amount)
	}
	if len(tx.TaxDetails) != 1 || tx.TaxDetails[0].Percent != 20.0 {
		t.Errorf("taxDetails = %+v, want [{percent:20}]", tx.TaxDetails)
	}
}

func TestReporter_Submit_CrossBorder(t *testing.T) {
	var gotBody transactionBody
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	r := newTestReporter(srv.URL)
	_, err := r.SubmitReport(context.Background(), pdp.SubmitReportInput{
		UserID:      "u1",
		IssuerSIREN: "123456789",
		Kind:        pdp.ReportCrossBorderB2C,
		Period:      pdp.ReportPeriod{Year: 2026, Month: 1},
		Lines:       []pdp.ReportLine{{TaxRate: "5.5", CountryCode: "DE", BaseHTCents: 5000, VATCents: 275}},
	})
	if err != nil {
		t.Fatalf("SubmitReport: %v", err)
	}
	if len(gotBody.Transactions) != 1 || gotBody.Transactions[0].CategoryCode != "TNT1" {
		t.Errorf("categoryCode = %q, want TNT1", gotBody.Transactions[0].CategoryCode)
	}
	if gotBody.TransactionDate != "2026-01-31" {
		t.Errorf("transactionDate = %q, want 2026-01-31", gotBody.TransactionDate)
	}
}

func TestReporter_Submit_Non202IsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"validationIssues":[{"statusMessage":"invalid"}]}`))
	}))
	defer srv.Close()

	r := newTestReporter(srv.URL)
	_, err := r.SubmitReport(context.Background(), pdp.SubmitReportInput{
		IssuerSIREN: "123456789",
		Lines:       []pdp.ReportLine{{TaxRate: "20"}},
	})
	if err == nil {
		t.Fatal("expected error on 400, got nil")
	}
}

func TestReporter_Submit_MissingSIRENIsError(t *testing.T) {
	r := newTestReporter("http://unused")
	_, err := r.SubmitReport(context.Background(), pdp.SubmitReportInput{
		IssuerSIREN: "",
		Lines:       []pdp.ReportLine{{TaxRate: "20"}},
	})
	if err == nil {
		t.Fatal("expected error when IssuerSIREN is empty")
	}
}

func TestReporter_FetchReportStatus_AlwaysUnknown(t *testing.T) {
	r := newTestReporter("http://unused")
	got, err := r.FetchReportStatus(context.Background(), "some-id")
	if err != nil {
		t.Fatalf("FetchReportStatus: %v", err)
	}
	if got != pdp.PlatformUnknown {
		t.Errorf("got %q, want UNKNOWN", got)
	}
}

func TestLastDayOfMonth(t *testing.T) {
	cases := []struct {
		p    pdp.ReportPeriod
		want string
	}{
		{pdp.ReportPeriod{Year: 2026, Month: 1}, "2026-01-31"},
		{pdp.ReportPeriod{Year: 2026, Month: 2}, "2026-02-28"},
		{pdp.ReportPeriod{Year: 2024, Month: 2}, "2024-02-29"}, // leap year
		{pdp.ReportPeriod{Year: 2026, Month: 12}, "2026-12-31"},
	}
	for _, tc := range cases {
		if got := lastDayOfMonth(tc.p); got != tc.want {
			t.Errorf("lastDayOfMonth(%v) = %q, want %q", tc.p, got, tc.want)
		}
	}
}
