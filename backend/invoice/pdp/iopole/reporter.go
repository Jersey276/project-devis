package iopole

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"project-devis-invoice/pdp"
)

// Reporter implements pdp.Reporter against the Iopole e-reporting API
// (POST /v1/reporting/transaction/scheme/0002/value/{siren}). One call per
// period covers the full monthly aggregate; transactionDate is set to the last
// day of the civil month. FetchReportStatus always returns PlatformUnknown
// because the Iopole GET endpoint is not yet stable.
type Reporter struct {
	http       *http.Client
	baseURL    string
	customerID string
	token      tokenProvider
}

// NewReporter builds the Iopole e-reporting adapter. It shares the same base
// URL, OAuth2 credentials and customer-id as the invoice Client/Directory.
func NewReporter(baseURL, tokenURL, clientID, clientSecret, customerID string) *Reporter {
	return &Reporter{
		http:       &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
		customerID: customerID,
		token:      newOAuthTokenProvider(clientID, clientSecret, tokenURL),
	}
}

func (r *Reporter) SubmitReport(ctx context.Context, in pdp.SubmitReportInput) (pdp.SubmitReportResult, error) {
	if in.IssuerSIREN == "" {
		return pdp.SubmitReportResult{}, fmt.Errorf("iopole reporter: IssuerSIREN is required")
	}

	body, err := buildTransactionBody(in)
	if err != nil {
		return pdp.SubmitReportResult{}, err
	}

	url := r.baseURL + "/v1/reporting/transaction/scheme/0002/value/" + in.IssuerSIREN
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return pdp.SubmitReportResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if err := authorize(ctx, req, r.token, r.customerID); err != nil {
		return pdp.SubmitReportResult{}, err
	}

	resp, err := r.http.Do(req)
	if err != nil {
		return pdp.SubmitReportResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return pdp.SubmitReportResult{}, fmt.Errorf("iopole e-reporting: status %d: %s", resp.StatusCode, readSnippet(resp.Body))
	}

	return pdp.SubmitReportResult{ReportID: "", Status: pdp.PlatformSubmitted}, nil
}

// FetchReportStatus always returns PlatformUnknown: the Iopole GET report
// endpoint is not yet stable, so the report poller leaves all rows untouched.
func (r *Reporter) FetchReportStatus(_ context.Context, _ string) (pdp.PlatformStatus, error) {
	return pdp.PlatformUnknown, nil
}

// categoryCode maps a ReportKind to its Iopole transaction category.
//   - TRANSACTION (B5) → TPS1: services subject to VAT (domestic FR B2C)
//   - CROSS_BORDER_B2C (C5) → TNT1: intra-EU distance sales, not subject to FR VAT
func categoryCode(kind pdp.ReportKind) string {
	if kind == pdp.ReportCrossBorderB2C {
		return "TNT1"
	}
	return "TPS1"
}

// lastDayOfMonth returns YYYY-MM-DD for the last calendar day of the period.
func lastDayOfMonth(p pdp.ReportPeriod) string {
	last := time.Date(p.Year, time.Month(p.Month)+1, 0, 0, 0, 0, 0, time.UTC)
	return last.Format("2006-01-02")
}

type transactionBody struct {
	TransactionDate string        `json:"transactionDate"`
	Transactions    []transaction `json:"transactions"`
}

type transaction struct {
	CategoryCode string   `json:"categoryCode"`
	Currency     string   `json:"currency"`
	Monetary     monetary `json:"monetary"`
	TaxDetails   []taxDetail `json:"taxDetails"`
}

type monetary struct {
	TaxBasisTotalAmount amount `json:"taxBasisTotalAmount"`
	TaxTotalAmount      amount `json:"taxTotalAmount"`
}

type taxDetail struct {
	TaxableAmount amount  `json:"taxableAmount"`
	TaxAmount     amount  `json:"taxAmount"`
	Percent       float64 `json:"percent"`
}

type amount struct {
	Amount float64 `json:"amount"`
}

func buildTransactionBody(in pdp.SubmitReportInput) ([]byte, error) {
	code := categoryCode(in.Kind)
	txs := make([]transaction, 0, len(in.Lines))
	for _, l := range in.Lines {
		rate, err := strconv.ParseFloat(l.TaxRate, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid tax rate %q: %w", l.TaxRate, err)
		}
		baseHT := float64(l.BaseHTCents) / 100
		vat := float64(l.VATCents) / 100
		txs = append(txs, transaction{
			CategoryCode: code,
			Currency:     "EUR",
			Monetary: monetary{
				TaxBasisTotalAmount: amount{Amount: baseHT},
				TaxTotalAmount:      amount{Amount: vat},
			},
			TaxDetails: []taxDetail{
				{
					TaxableAmount: amount{Amount: baseHT},
					TaxAmount:     amount{Amount: vat},
					Percent:       rate,
				},
			},
		})
	}
	return json.Marshal(transactionBody{
		TransactionDate: lastDayOfMonth(in.Period),
		Transactions:    txs,
	})
}
