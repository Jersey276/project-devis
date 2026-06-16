// Package oauth implements the gateway-side Authorization Code flow for the
// configured identity providers (Google, GitHub, Microsoft). The gateway is the
// only HTTP service, so it handles the redirect + callback, exchanges the code,
// fetches the user's email + stable id, and hands a verified identity to the
// auth service via the OAuthLogin RPC.
package oauth

import "context"

// UserInfo is the normalized identity returned by every provider.
type UserInfo struct {
	Email    string
	Sub      string // stable, provider-specific identifier
	Verified bool   // whether the provider attests the email is verified
}

// Provider abstracts a single OAuth2 identity provider.
type Provider interface {
	// Name returns the provider key ("google" | "github" | "microsoft").
	Name() string
	// AuthCodeURL builds the consent-screen URL carrying the CSRF state.
	AuthCodeURL(state string) string
	// Exchange swaps the authorization code for a token and fetches userinfo.
	Exchange(ctx context.Context, code string) (UserInfo, error)
}
