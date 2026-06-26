package line

import (
	"testing"

	"project-devis-quote/actions/sqlutil"
)

func TestValidateData_Simple(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"empty string normalises to default line kind", "", `{"kind":"line"}`, false},
		{"null normalises to default line kind", "null", `{"kind":"line"}`, false},
		{"empty object accepted", "{}", `{"kind":"line"}`, false},
		{"text line accepted", `{"kind":"text","description":"Note"}`, `{"kind":"text","description":"Note"}`, false},
		{"fee line accepted", `{"kind":"fee","fee_id":"f-1"}`, `{"kind":"fee","fee_id":"f-1"}`, false},
		{"fee line without fee_id rejected", `{"kind":"fee"}`, "", true},
		{"non-empty object rejected", `{"foo":"bar"}`, "", true},
		{"invalid JSON rejected", `not json`, "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ValidateData(sqlutil.TypeSimple, c.input)
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
		{"valid with optional unit", `{"sublines":[{"name":"a","quantity":"1.5","unit":"kg","unit_price":1000,"option":true}]}`, false},
		{"subline referencing a fee accepted", `{"sublines":[{"name":"a","quantity":"1","unit_price":1000,"fee_id":"f-1"}]}`, false},
		{"empty data rejected", "", true},
		{"empty sublines accepted", `{"sublines":[]}`, false},
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
			_, err := ValidateData(sqlutil.TypeMultiple, c.input)
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

func TestFeeIDFromData(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"fee line returns its fee_id", `{"kind":"fee","fee_id":"f-1"}`, "f-1"},
		{"non-fee line returns empty", `{"kind":"line"}`, ""},
		{"text line returns empty", `{"kind":"text","description":"x"}`, ""},
		{"empty data returns empty", "", ""},
		{"invalid JSON returns empty", "not json", ""},
		{"fee_id without fee kind ignored", `{"kind":"line","fee_id":"f-1"}`, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := FeeIDFromData(c.input); got != c.want {
				t.Fatalf("expected %q, got %q", c.want, got)
			}
		})
	}
}
