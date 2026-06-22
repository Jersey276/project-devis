package iopole

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// tokenProvider yields a bearer token for an outbound request. It is an interface
// so tests inject a static token and never reach the real Keycloak.
type tokenProvider interface {
	Token(ctx context.Context) (string, error)
}

// oauthTokenProvider gets tokens via OAuth2 client_credentials (Iopole = Keycloak).
// The underlying TokenSource caches the token and refreshes it before expiry.
type oauthTokenProvider struct {
	src oauth2.TokenSource
}

func newOAuthTokenProvider(clientID, clientSecret, tokenURL string) *oauthTokenProvider {
	cfg := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
	}
	return &oauthTokenProvider{src: cfg.TokenSource(context.Background())}
}

func (p *oauthTokenProvider) Token(ctx context.Context) (string, error) {
	t, err := p.src.Token()
	if err != nil {
		return "", err
	}
	return t.AccessToken, nil
}

// staticTokenProvider returns a fixed token; used by tests.
type staticTokenProvider string

func (s staticTokenProvider) Token(context.Context) (string, error) { return string(s), nil }

// authorize sets the bearer token and tenant header common to every Iopole call.
func authorize(ctx context.Context, req *http.Request, tp tokenProvider, customerID string) error {
	tok, err := tp.Token(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	if customerID != "" {
		req.Header.Set("customer-id", customerID)
	}
	return nil
}
