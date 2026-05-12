// Package gotenberg is a thin HTTP client over the Gotenberg Chromium HTML→PDF
// endpoint. It speaks multipart/form-data and returns the raw PDF bytes.
package gotenberg

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"
)

// maxPDFBytes caps the PDF we'll read from Gotenberg; mirrors the gRPC
// message cap so we can't accept something the next hop will reject.
const maxPDFBytes = 8 * 1024 * 1024

type Client struct {
	addr       string
	httpClient *http.Client
}

// New: timeout is set just above Gotenberg's own --api-timeout (30s in
// docker-compose) so Gotenberg's structured error response wins the race
// instead of a generic Go context-deadline.
func New(addr string) *Client {
	return &Client{
		addr:       addr,
		httpClient: &http.Client{Timeout: 45 * time.Second},
	}
}

// Convert posts the given HTML document to Gotenberg's Chromium HTML→PDF
// endpoint and returns the rendered PDF bytes. Gotenberg requires the entry
// file to be named exactly "index.html".
func (c *Client) Convert(ctx context.Context, html []byte) ([]byte, error) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="files"; filename="index.html"`)
	h.Set("Content-Type", "text/html")
	fw, err := mw.CreatePart(h)
	if err != nil {
		return nil, fmt.Errorf("gotenberg: build multipart: %w", err)
	}
	if _, err := fw.Write(html); err != nil {
		return nil, fmt.Errorf("gotenberg: write html: %w", err)
	}
	if err := mw.Close(); err != nil {
		return nil, fmt.Errorf("gotenberg: close multipart: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.addr+"/forms/chromium/convert/html", &body)
	if err != nil {
		return nil, fmt.Errorf("gotenberg: build request: %w", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gotenberg: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Gotenberg returns text/plain explanations on 4xx/5xx — useful for
		// debugging Chromium crashes. Cap the body to avoid unbounded reads.
		excerpt, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("gotenberg %d: %s", resp.StatusCode, bytes.TrimSpace(excerpt))
	}

	pdf, err := io.ReadAll(io.LimitReader(resp.Body, maxPDFBytes+1))
	if err != nil {
		return nil, fmt.Errorf("gotenberg: read body: %w", err)
	}
	if len(pdf) > maxPDFBytes {
		return nil, fmt.Errorf("gotenberg: PDF exceeds %d bytes", maxPDFBytes)
	}
	return pdf, nil
}
