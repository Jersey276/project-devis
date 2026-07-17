package middleware

import (
	"encoding/json"
	"testing"
)

func TestRedactJSONBody(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  map[string]any
	}{
		{
			name:  "login password",
			input: `{"email":"a@b.com","password":"hunter2"}`,
			want:  map[string]any{"email": "a@b.com", "password": redactedValue},
		},
		{
			name:  "password update",
			input: `{"old_password":"old","new_password":"new"}`,
			want:  map[string]any{"old_password": redactedValue, "new_password": redactedValue},
		},
		{
			name:  "bank details case-insensitive",
			input: `{"IBAN":"FR7630006000011234567890189","Bic":"AGRIFRPP"}`,
			want:  map[string]any{"IBAN": redactedValue, "Bic": redactedValue},
		},
		{
			name:  "nested object",
			input: `{"user":{"email":"a@b.com","iban":"FR76..."}}`,
			want: map[string]any{"user": map[string]any{
				"email": "a@b.com",
				"iban":  redactedValue,
			}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := redactJSONBody(tc.input)

			var gotParsed map[string]any
			if err := json.Unmarshal([]byte(got), &gotParsed); err != nil {
				t.Fatalf("redacted body is not valid JSON: %v (body=%s)", err, got)
			}

			wantJSON, _ := json.Marshal(tc.want)
			var wantParsed map[string]any
			_ = json.Unmarshal(wantJSON, &wantParsed)

			gotStr, _ := json.Marshal(gotParsed)
			wantStr, _ := json.Marshal(wantParsed)
			if string(gotStr) != string(wantStr) {
				t.Errorf("got %s, want %s", gotStr, wantStr)
			}
		})
	}
}

func TestRedactJSONBody_NonJSONPassthrough(t *testing.T) {
	input := "page=1&limit=20"
	if got := redactJSONBody(input); got != input {
		t.Errorf("expected non-JSON query string to pass through unchanged, got %q", got)
	}
}

func TestRedactJSONBody_ArrayOfObjects(t *testing.T) {
	input := `[{"email":"a@b.com","password":"p1"},{"email":"c@d.com","password":"p2"}]`
	got := redactJSONBody(input)

	var arr []map[string]any
	if err := json.Unmarshal([]byte(got), &arr); err != nil {
		t.Fatalf("redacted body is not valid JSON array: %v", err)
	}
	for i, obj := range arr {
		if obj["password"] != redactedValue {
			t.Errorf("element %d: expected password redacted, got %v", i, obj["password"])
		}
	}
}
