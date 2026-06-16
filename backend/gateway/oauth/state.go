package oauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

// Mode distinguishes the two OAuth round-trip intents carried in the state.
const (
	ModeLogin = "login" // public flow: authenticate or provision a user
	ModeLink  = "link"  // authenticated flow: attach a provider to the caller
)

// State is the CSRF/state payload carried through the OAuth round-trip. It is
// HMAC-signed and stored both in a cookie and in the provider's `state` query
// param (double-submit). The nonce defeats replay; next is the post-login
// landing path. Mode/UserID carry the link-account context (UserID is empty for
// the login flow). Because the payload is HMAC-signed, the user cannot forge a
// UserID — the gateway sets it from the caller's JWT in OAuthBegin.
type State struct {
	Nonce  string
	Next   string
	Mode   string
	UserID string
}

// NewState creates a login-mode State with a fresh random nonce and a validated
// next path.
func NewState(next string) State {
	return State{Nonce: randomNonce(), Next: SafeNextPath(next), Mode: ModeLogin}
}

// NewLinkState creates a link-mode State carrying the authenticated user's id.
func NewLinkState(next, userID string) State {
	return State{Nonce: randomNonce(), Next: SafeNextPath(next), Mode: ModeLink, UserID: userID}
}

// SafeNextPath mirrors the frontend safeNextPath: only same-origin absolute
// paths are allowed, defaulting to "/quote" otherwise.
func SafeNextPath(value string) string {
	if value == "" || !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") {
		return "/quote"
	}
	return value
}

// Sign serializes and HMAC-signs the state with the given secret, returning a
// compact token "<base64(nonce|mode|userID|next)>.<hex(hmac)>". Next is last so
// it may safely contain the separator.
func (s State) Sign(secret string) string {
	mode := s.Mode
	if mode == "" {
		mode = ModeLogin
	}
	payload := base64.RawURLEncoding.EncodeToString(
		[]byte(s.Nonce + "|" + mode + "|" + s.UserID + "|" + s.Next),
	)
	return payload + "." + sign(payload, secret)
}

// ParseState verifies the token's signature and returns the decoded State.
// It returns ok=false on any tampering or malformed input. Two-segment payloads
// from before the link feature are still accepted as login-mode states.
func ParseState(token, secret string) (State, bool) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return State{}, false
	}
	payload, providedMAC := parts[0], parts[1]
	expectedMAC := sign(payload, secret)
	if !hmac.Equal([]byte(providedMAC), []byte(expectedMAC)) {
		return State{}, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return State{}, false
	}
	segments := strings.SplitN(string(raw), "|", 4)
	switch len(segments) {
	case 2: // legacy "nonce|next" → login mode
		return State{Nonce: segments[0], Next: SafeNextPath(segments[1]), Mode: ModeLogin}, true
	case 4:
		return State{
			Nonce:  segments[0],
			Mode:   segments[1],
			UserID: segments[2],
			Next:   SafeNextPath(segments[3]),
		}, true
	default:
		return State{}, false
	}
}

func sign(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func randomNonce() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
