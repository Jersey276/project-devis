package oauth

import "testing"

func TestState_SignParseRoundTrip(t *testing.T) {
	secret := "test-secret"
	s := NewState("/quote/42")
	token := s.Sign(secret)

	got, ok := ParseState(token, secret)
	if !ok {
		t.Fatal("expected valid state")
	}
	if got.Nonce != s.Nonce || got.Next != "/quote/42" {
		t.Fatalf("round-trip mismatch: got %+v want %+v", got, s)
	}
}

func TestState_TamperRejected(t *testing.T) {
	secret := "test-secret"
	token := NewState("/quote").Sign(secret)

	if _, ok := ParseState(token+"x", secret); ok {
		t.Fatal("tampered signature should be rejected")
	}
	if _, ok := ParseState(token, "other-secret"); ok {
		t.Fatal("wrong secret should be rejected")
	}
	if _, ok := ParseState("garbage", secret); ok {
		t.Fatal("malformed token should be rejected")
	}
}

func TestSafeNextPath(t *testing.T) {
	cases := map[string]string{
		"/quote":          "/quote",
		"/quote/1?x=2":    "/quote/1?x=2",
		"":                "/quote",
		"//evil.com":      "/quote",
		"https://evil.com": "/quote",
		"relative":        "/quote",
	}
	for in, want := range cases {
		if got := SafeNextPath(in); got != want {
			t.Errorf("SafeNextPath(%q) = %q, want %q", in, got, want)
		}
	}
}
