package line

import "testing"

func TestValidateData_Simple(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"empty string normalises to {}", "", "{}", false},
		{"null normalises to {}", "null", "{}", false},
		{"empty object accepted", "{}", "{}", false},
		{"non-empty object rejected", `{"foo":"bar"}`, "", true},
		{"invalid JSON rejected", `not json`, "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ValidateData(TypeSimple, c.input)
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil", c.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Fatalf("expected %q, got %q", c.want, got)
			}
		})
	}
}

func TestValidateData_Multiple(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid sublines", `{"sublines":[{"name":"a","quantity":"1","unit_price":1000}]}`, false},
		{"valid with optional unit", `{"sublines":[{"name":"a","quantity":"1.5","unit":"kg","unit_price":1000}]}`, false},
		{"empty data rejected", "", true},
		{"empty sublines rejected", `{"sublines":[]}`, true},
		{"missing name rejected", `{"sublines":[{"quantity":"1","unit_price":1}]}`, true},
		{"missing quantity rejected", `{"sublines":[{"name":"a","unit_price":1}]}`, true},
		{"non-numeric quantity rejected", `{"sublines":[{"name":"a","quantity":"abc","unit_price":1}]}`, true},
		{"negative unit_price rejected", `{"sublines":[{"name":"a","quantity":"1","unit_price":-1}]}`, true},
		{"unknown field rejected", `{"sublines":[{"name":"a","quantity":"1","unit_price":1,"foo":"bar"}]}`, true},
		{"top-level unknown field rejected", `{"sublines":[{"name":"a","quantity":"1","unit_price":1}],"foo":"bar"}`, true},
		{"invalid JSON rejected", `not json`, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ValidateData(TypeMultiple, c.input)
			if c.wantErr && err == nil {
				t.Fatalf("expected error, got nil for input %q", c.input)
			}
			if !c.wantErr && err != nil {
				t.Fatalf("unexpected error for input %q: %v", c.input, err)
			}
		})
	}
}

func TestValidateData_UnknownType(t *testing.T) {
	if _, err := ValidateData("weird", "{}"); err == nil {
		t.Fatal("expected error for unknown type")
	}
}
