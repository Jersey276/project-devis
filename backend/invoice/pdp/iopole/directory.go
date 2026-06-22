package iopole

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"project-devis-invoice/pdp"
)

// Directory implements pdp.Directory against the Iopole French directory
// (GET /v1/directory/french?q={siret}).
type Directory struct {
	http       *http.Client
	baseURL    string
	customerID string
	token      tokenProvider
}

// NewDirectory builds the Iopole directory adapter, sharing the same auth as Client.
func NewDirectory(baseURL, tokenURL, clientID, clientSecret, customerID string) *Directory {
	return &Directory{
		http:       &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
		customerID: customerID,
		token:      newOAuthTokenProvider(clientID, clientSecret, tokenURL),
	}
}

// Resolve looks the recipient up in the French directory. Iopole's /v1/directory/french
// accepts a SIREN (9 digits); if a full SIRET (14 digits) is provided we strip the
// NIC suffix. An empty result set means the recipient is not reachable for e-invoicing
// → pdp.ErrRecipientNotFound, which the deposit flow maps to RecipientNotInDirectory (4014).
func (d *Directory) Resolve(ctx context.Context, siret string) (pdp.RecipientRouting, error) {
	q := siret
	if len(siret) == 14 {
		q = siret[:9] // Iopole rejects 14-digit queries; search by SIREN only
	}
	u := fmt.Sprintf("%s/v1/directory/french?q=%s",
		d.baseURL, url.QueryEscape(q))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return pdp.RecipientRouting{}, err
	}
	if err := authorize(ctx, req, d.token, d.customerID); err != nil {
		return pdp.RecipientRouting{}, err
	}

	resp, err := d.http.Do(req)
	if err != nil {
		return pdp.RecipientRouting{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return pdp.RecipientRouting{}, fmt.Errorf("iopole directory: status %d: %s", resp.StatusCode, readSnippet(resp.Body))
	}

	var out struct {
		Data []struct {
			BusinessEntityID string `json:"businessEntityId"`
			Name             string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return pdp.RecipientRouting{}, fmt.Errorf("decode directory response: %w", err)
	}
	if len(out.Data) == 0 {
		return pdp.RecipientRouting{}, pdp.ErrRecipientNotFound
	}
	return pdp.RecipientRouting{
		RoutingID:    out.Data[0].BusinessEntityID,
		PlatformName: out.Data[0].Name,
	}, nil
}
