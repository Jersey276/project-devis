// Package iopole is the real Plateforme Agréée adapter (Iopole) behind the neutral
// pdp.Client / pdp.Directory seams. It owns the HTTP transport, OAuth2 auth and the
// Factur-X document fetch, so the pdp package itself stays I/O-free. Selected at
// startup via PDP_PROVIDER=iopole; the no-op adapters remain the default.
package iopole

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sort"
	"time"

	"project-devis-invoice/pdp"
)

// DocumentSource yields the Factur-X PDF/A-3 bytes to deposit for one issued
// invoice. The invoice service implements it over export.ExportInvoice(facturx=true);
// keeping it an interface lets the adapter stay free of gRPC and lets tests stub it.
type DocumentSource interface {
	FetchFacturX(ctx context.Context, invoiceID, userID string) ([]byte, error)
}

// Client implements pdp.Client against the Iopole API.
type Client struct {
	http       *http.Client
	baseURL    string // e.g. https://api.ppd.iopole.fr/v1/api
	customerID string
	token      tokenProvider
	docs       DocumentSource
}

// NewClient builds the Iopole platform client. tokenURL/clientID/clientSecret drive
// OAuth2 client_credentials; customerID is the Iopole tenant header; docs supplies
// the Factur-X bytes.
func NewClient(baseURL, tokenURL, clientID, clientSecret, customerID string, docs DocumentSource) *Client {
	return &Client{
		http:       &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
		customerID: customerID,
		token:      newOAuthTokenProvider(clientID, clientSecret, tokenURL),
		docs:       docs,
	}
}

// Submit fetches the Factur-X PDF/A-3 and deposits it on Iopole (POST /v1/invoice,
// multipart field "file"). On 201 it returns the platform-assigned id as the
// submission handle.
func (c *Client) Submit(ctx context.Context, in pdp.SubmitInput) (pdp.SubmitResult, error) {
	pdfBytes, err := c.docs.FetchFacturX(ctx, in.InvoiceID, in.UserID)
	if err != nil {
		return pdp.SubmitResult{}, fmt.Errorf("fetch factur-x: %w", err)
	}

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", in.InvoiceNumber+".pdf")
	if err != nil {
		return pdp.SubmitResult{}, err
	}
	if _, err := part.Write(pdfBytes); err != nil {
		return pdp.SubmitResult{}, err
	}
	if err := w.Close(); err != nil {
		return pdp.SubmitResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/invoice", &body)
	if err != nil {
		return pdp.SubmitResult{}, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	if err := authorize(ctx, req, c.token, c.customerID); err != nil {
		return pdp.SubmitResult{}, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return pdp.SubmitResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return pdp.SubmitResult{}, fmt.Errorf("iopole deposit: status %d: %s", resp.StatusCode, readSnippet(resp.Body))
	}

	var created struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return pdp.SubmitResult{}, fmt.Errorf("decode deposit response: %w", err)
	}
	if created.ID == "" {
		return pdp.SubmitResult{}, fmt.Errorf("iopole deposit: empty id in response")
	}
	return pdp.SubmitResult{SubmissionID: created.ID, Status: pdp.PlatformSubmitted}, nil
}

// FetchStatus reads the status history (GET /v1/invoice/{id}/status-history) and maps
// the latest status code to a PlatformStatus. An empty history or an unmapped code
// yields PlatformUnknown so the poller leaves the invoice untouched.
func (c *Client) FetchStatus(ctx context.Context, submissionID string) (pdp.PlatformStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/v1/invoice/"+submissionID+"/status-history", nil)
	if err != nil {
		return pdp.PlatformUnknown, err
	}
	if err := authorize(ctx, req, c.token, c.customerID); err != nil {
		return pdp.PlatformUnknown, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return pdp.PlatformUnknown, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return pdp.PlatformUnknown, fmt.Errorf("iopole status: status %d: %s", resp.StatusCode, readSnippet(resp.Body))
	}

	var items []statusItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return pdp.PlatformUnknown, fmt.Errorf("decode status response: %w", err)
	}
	if len(items) == 0 {
		return pdp.PlatformUnknown, nil
	}
	// History is not guaranteed ordered; take the most recent by date.
	sort.SliceStable(items, func(i, j int) bool { return items[i].Date.After(items[j].Date) })
	return mapStatus(items[0].Status.Code), nil
}

type statusItem struct {
	Date   time.Time `json:"date"`
	Status struct {
		Code string `json:"code"`
	} `json:"status"`
}

// readSnippet reads a bounded chunk of an error body for diagnostics without
// risking an unbounded read.
func readSnippet(r io.Reader) string {
	b, _ := io.ReadAll(io.LimitReader(r, 512))
	return string(b)
}
