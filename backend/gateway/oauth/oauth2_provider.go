package oauth

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

// fetchFunc retrieves and normalizes the provider's userinfo using an
// authenticated HTTP client built from the exchanged token.
type fetchFunc func(ctx context.Context, client *http.Client) (UserInfo, error)

// oauth2Provider is a generic Provider backed by an *oauth2.Config plus a
// provider-specific userinfo fetcher (each provider's endpoint + JSON differs).
type oauth2Provider struct {
	name  string
	cfg   *oauth2.Config
	fetch fetchFunc
}

func (p *oauth2Provider) Name() string { return p.name }

func (p *oauth2Provider) AuthCodeURL(state string) string {
	return p.cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (p *oauth2Provider) Exchange(ctx context.Context, code string) (UserInfo, error) {
	token, err := p.cfg.Exchange(ctx, code)
	if err != nil {
		return UserInfo{}, err
	}
	client := p.cfg.Client(ctx, token)
	return p.fetch(ctx, client)
}
