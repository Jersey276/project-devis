package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

// Userinfo endpoints. Exposed as vars so tests can point them at an httptest server.
var (
	googleUserinfoURL    = "https://www.googleapis.com/oauth2/v3/userinfo"
	githubUserURL        = "https://api.github.com/user"
	githubUserEmailsURL  = "https://api.github.com/user/emails"
	microsoftUserinfoURL = "https://graph.microsoft.com/v1.0/me"
)

// providerSpecs is the single source of truth for supported providers and their
// OAuth2 wiring. Adding a provider is one entry here plus its fetch function.
var providerSpecs = []struct {
	name     string
	envKey   string
	endpoint oauth2.Endpoint
	scopes   []string
	fetch    fetchFunc
}{
	{"google", "GOOGLE", google.Endpoint, []string{"openid", "email"}, fetchGoogle},
	{"github", "GITHUB", github.Endpoint, []string{"read:user", "user:email"}, fetchGitHub},
	{"microsoft", "MICROSOFT", microsoft.AzureADEndpoint("common"), []string{"openid", "email", "User.Read"}, fetchMicrosoft},
}

// FromEnv builds the set of providers whose client id/secret/redirect env vars
// are all set. Providers without complete configuration are simply omitted.
func FromEnv() map[string]Provider {
	providers := map[string]Provider{}
	for _, spec := range providerSpecs {
		cfg := envConfig(spec.envKey, spec.endpoint, spec.scopes)
		if cfg == nil {
			continue
		}
		providers[spec.name] = &oauth2Provider{name: spec.name, cfg: cfg, fetch: spec.fetch}
	}
	return providers
}

func envConfig(prefix string, endpoint oauth2.Endpoint, scopes []string) *oauth2.Config {
	id := os.Getenv("OAUTH_" + prefix + "_CLIENT_ID")
	secret := os.Getenv("OAUTH_" + prefix + "_CLIENT_SECRET")
	redirect := os.Getenv("OAUTH_" + prefix + "_REDIRECT_URL")
	if id == "" || secret == "" || redirect == "" {
		return nil
	}
	return &oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		RedirectURL:  redirect,
		Endpoint:     endpoint,
		Scopes:       scopes,
	}
}

func fetchGoogle(ctx context.Context, client *http.Client) (UserInfo, error) {
	var body struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := getJSON(ctx, client, googleUserinfoURL, &body); err != nil {
		return UserInfo{}, err
	}
	if body.Sub == "" || body.Email == "" {
		return UserInfo{}, fmt.Errorf("google: incomplete userinfo")
	}
	return UserInfo{Email: body.Email, Sub: body.Sub, Verified: body.EmailVerified}, nil
}

func fetchGitHub(ctx context.Context, client *http.Client) (UserInfo, error) {
	var user struct {
		ID    int64  `json:"id"`
		Email string `json:"email"`
	}
	if err := getJSON(ctx, client, githubUserURL, &user); err != nil {
		return UserInfo{}, err
	}
	if user.ID == 0 {
		return UserInfo{}, fmt.Errorf("github: missing user id")
	}

	// /user may omit the email; the verified primary email lives on /user/emails.
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := getJSON(ctx, client, githubUserEmailsURL, &emails); err != nil {
		return UserInfo{}, err
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return UserInfo{Email: e.Email, Sub: fmt.Sprintf("%d", user.ID), Verified: true}, nil
		}
	}
	return UserInfo{}, fmt.Errorf("github: no verified primary email")
}

func fetchMicrosoft(ctx context.Context, client *http.Client) (UserInfo, error) {
	var body struct {
		ID                string `json:"id"`
		Mail              string `json:"mail"`
		UserPrincipalName string `json:"userPrincipalName"`
	}
	if err := getJSON(ctx, client, microsoftUserinfoURL, &body); err != nil {
		return UserInfo{}, err
	}
	email := body.Mail
	if email == "" {
		email = body.UserPrincipalName
	}
	if body.ID == "" || email == "" {
		return UserInfo{}, fmt.Errorf("microsoft: incomplete userinfo")
	}
	// Microsoft work/school accounts are organization-verified.
	return UserInfo{Email: email, Sub: body.ID, Verified: true}, nil
}

func getJSON(ctx context.Context, client *http.Client, url string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("userinfo request to %s failed: %s", url, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
