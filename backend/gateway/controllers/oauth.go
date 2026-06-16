package controllers

import (
	"net/http"
	"net/url"
	"os"

	auth "gateway/auth"
	"gateway/middleware"
	"gateway/oauth"

	"github.com/gin-gonic/gin"
)

const oauthStateCookie = "oauth-state"

// oauthStateMaxAge bounds how long a consent round-trip may take.
const oauthStateMaxAge = 600

func oauthStateSecret() string { return os.Getenv("OAUTH_STATE_SECRET") }

func oauthFailureRedirect() string {
	if v := os.Getenv("OAUTH_FAILURE_REDIRECT"); v != "" {
		return v
	}
	return "/login"
}

// failOAuth redirects the browser to the failure page with a short error slug.
func failOAuth(c *gin.Context, slug string) {
	redirectWithQuery(c, oauthFailureRedirect(), "oauth_error", slug)
}

// OAuthBegin starts the public Authorization Code flow: it sets a signed state
// cookie and redirects to the provider's consent screen.
func OAuthBegin(c *gin.Context, providers map[string]oauth.Provider) {
	beginOAuth(c, providers, oauth.NewState(c.Query("next")))
}

// OAuthLinkBegin starts the link flow for an already-authenticated user. The
// caller's user_id is read from the JWT context and embedded in the signed
// state, so the callback can attach the identity to the right account without
// trusting any client-supplied value.
func OAuthLinkBegin(c *gin.Context, providers map[string]oauth.Provider) {
	beginOAuth(c, providers, oauth.NewLinkState(c.Query("next"), c.GetString(middleware.CtxUserID)))
}

// beginOAuth sets the signed state cookie and redirects to the consent screen.
func beginOAuth(c *gin.Context, providers map[string]oauth.Provider, state oauth.State) {
	provider, ok := providers[c.Param("provider")]
	if !ok {
		failOAuth(c, "unknown_provider")
		return
	}

	signed := state.Sign(oauthStateSecret())

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oauthStateCookie, signed, oauthStateMaxAge, "/", "", cookieSecure, true)
	c.Redirect(http.StatusFound, provider.AuthCodeURL(signed))
}

// OAuthCallback validates the CSRF state, exchanges the code, provisions/links
// the account via the auth service, sets the auth cookies, and redirects to the
// post-login landing path.
func OAuthCallback(c *gin.Context, providers map[string]oauth.Provider, client auth.AuthServiceClient) {
	provider, ok := providers[c.Param("provider")]
	if !ok {
		failOAuth(c, "unknown_provider")
		return
	}

	// CSRF double-submit: the query state must match the cookie and verify.
	cookie, err := c.Cookie(oauthStateCookie)
	c.SetCookie(oauthStateCookie, "", -1, "/", "", cookieSecure, true) // single-use
	queryState := c.Query("state")
	if err != nil || queryState == "" || queryState != cookie {
		failOAuth(c, "state")
		return
	}
	state, valid := oauth.ParseState(queryState, oauthStateSecret())
	if !valid {
		failOAuth(c, "state")
		return
	}

	code := c.Query("code")
	if code == "" {
		failOAuth(c, "provider")
		return
	}

	info, err := provider.Exchange(c.Request.Context(), code)
	if err != nil {
		failOAuth(c, "provider")
		return
	}

	if state.Mode == oauth.ModeLink {
		oauthLinkCallback(c, provider, info, state, client)
		return
	}

	resp, err := client.OAuthLogin(c.Request.Context(), &auth.OAuthLoginRequest{
		Provider:       provider.Name(),
		ProviderUserId: info.Sub,
		Email:          info.Email,
		EmailVerified:  info.Verified,
	})
	if err != nil {
		failOAuth(c, "internal")
		return
	}
	if !resp.Success {
		switch resp.GetCode() {
		case CodeOAuthEmailNotVerified:
			failOAuth(c, "email_not_verified")
		default:
			failOAuth(c, "internal")
		}
		return
	}

	setAuthCookies(c, resp.GetToken(), resp.GetRefreshToken(), resp.GetRememberMe())
	c.Redirect(http.StatusFound, oauth.SafeNextPath(state.Next))
}

// oauthLinkCallback attaches the verified identity to the user carried in the
// signed state. No auth cookies are issued (the user is already logged in); it
// redirects back to the next path with an outcome query param.
func oauthLinkCallback(c *gin.Context, provider oauth.Provider, info oauth.UserInfo, state oauth.State, client auth.AuthServiceClient) {
	next := oauth.SafeNextPath(state.Next)
	if state.UserID == "" {
		redirectWithQuery(c, next, "oauth_error", "internal")
		return
	}

	resp, err := client.LinkOAuthIdentity(c.Request.Context(), &auth.LinkOAuthIdentityRequest{
		UserId:         state.UserID,
		Provider:       provider.Name(),
		ProviderUserId: info.Sub,
		Email:          info.Email,
		EmailVerified:  info.Verified,
	})
	if err != nil {
		redirectWithQuery(c, next, "oauth_error", "internal")
		return
	}
	if !resp.GetSuccess() {
		switch resp.GetCode() {
		case CodeOAuthIdentityTaken:
			redirectWithQuery(c, next, "oauth_error", "identity_taken")
		case CodeOAuthEmailNotVerified:
			redirectWithQuery(c, next, "oauth_error", "email_not_verified")
		default:
			redirectWithQuery(c, next, "oauth_error", "internal")
		}
		return
	}

	redirectWithQuery(c, next, "oauth_linked", provider.Name())
}

// redirectWithQuery appends key=value to path (preserving any existing query)
// and issues a 302.
func redirectWithQuery(c *gin.Context, path, key, value string) {
	sep := "?"
	if u, err := url.Parse(path); err == nil && u.RawQuery != "" {
		sep = "&"
	}
	c.Redirect(http.StatusFound, path+sep+key+"="+url.QueryEscape(value))
}

// ListOAuthIdentities returns the providers linked to the authenticated user.
func ListOAuthIdentities(c *gin.Context, client auth.AuthServiceClient) {
	resp, err := client.ListOAuthIdentities(c.Request.Context(), &auth.ListOAuthIdentitiesRequest{
		UserId: c.GetString(middleware.CtxUserID),
	})
	if err != nil {
		authErrors.unavailable(c)
		return
	}
	if !resp.GetSuccess() {
		authErrors.reply(c, resp.GetCode())
		return
	}

	providers := make([]gin.H, 0, len(resp.GetIdentities()))
	for _, ident := range resp.GetIdentities() {
		providers = append(providers, gin.H{"provider": ident.GetProvider(), "email": ident.GetEmail()})
	}
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"identities":   providers,
		"has_password": resp.GetHasPassword(),
	})
}

// UnlinkOAuthIdentity removes a linked provider from the authenticated user.
func UnlinkOAuthIdentity(c *gin.Context, client auth.AuthServiceClient) {
	resp, err := client.UnlinkOAuthIdentity(c.Request.Context(), &auth.UnlinkOAuthIdentityRequest{
		UserId:   c.GetString(middleware.CtxUserID),
		Provider: c.Param("provider"),
	})
	if err != nil {
		authErrors.unavailable(c)
		return
	}
	if !resp.GetSuccess() {
		authErrors.reply(c, resp.GetCode())
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
