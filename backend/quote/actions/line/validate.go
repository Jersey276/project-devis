package line

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"project-devis-quote/actions/sqlutil"
)

type lineDataPayload struct {
	Kind         string    `json:"kind,omitempty"`
	Description  string    `json:"description,omitempty"`
	Option       *bool     `json:"option,omitempty"`
	ParentLineID string    `json:"parent_line_id,omitempty"`
	FeeID        string    `json:"fee_id,omitempty"`
	Sublines     []subline `json:"sublines,omitempty"`
}

// ValidateData ensures the JSON payload matches the schema for the given line type.
// Returns the canonical JSON (defaulting Simple to "{}") if valid.
func ValidateData(lineType, data string) (string, error) {
	switch lineType {
	case sqlutil.TypeSimple:
		return validateSimple(data)
	case sqlutil.TypeMultiple:
		return validateMultiple(data)
	default:
		return "", fmt.Errorf("unknown line type %q", lineType)
	}
}

func validateSimple(data string) (string, error) {
	if data == "" || data == "null" {
		return `{"kind":"line"}`, nil
	}
	var payload lineDataPayload
	dec := json.NewDecoder(strings.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&payload); err != nil {
		return "", fmt.Errorf("simple line data invalid: %w", err)
	}
	if payload.Kind == "" {
		payload.Kind = "line"
	}
	if payload.Kind != "line" && payload.Kind != "text" && payload.Kind != "group" && payload.Kind != "subline" && payload.Kind != "fee" {
		return "", fmt.Errorf("simple line data has invalid kind %q", payload.Kind)
	}
	// Description can be empty on creation; it is filled in by the user later.
	if payload.Kind == "subline" && payload.ParentLineID == "" {
		return "", fmt.Errorf("subline parent_line_id is required")
	}
	// A fee line is a live reference to a catalog entry: fee_id is mandatory so
	// that updating the fee can propagate its snapshot back to this line.
	if payload.Kind == "fee" && payload.FeeID == "" {
		return "", fmt.Errorf("fee line fee_id is required")
	}
	clean, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(clean), nil
}

// FeeIDFromData returns the catalog fee_id carried by a top-level fee line
// (kind="fee"), or "" for any other line. It is used to populate the
// quote_lines.fee_id column, which the index/propagation query relies on.
// Sublines reference fees inside the JSON only, so they are not returned here.
func FeeIDFromData(data string) string {
	if data == "" {
		return ""
	}
	var payload lineDataPayload
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return ""
	}
	if payload.Kind != "fee" {
		return ""
	}
	return payload.FeeID
}

type subline struct {
	Name      string  `json:"name"`
	Quantity  string  `json:"quantity"`
	Unit      *string `json:"unit,omitempty"`
	UnitPrice int64   `json:"unit_price"`
	Option    *bool   `json:"option,omitempty"`
	FeeID     string  `json:"fee_id,omitempty"`
}

type multiplePayload struct {
	Kind        string    `json:"kind,omitempty"`
	Description string    `json:"description,omitempty"`
	Sublines    []subline `json:"sublines,omitempty"`
}

func validateMultiple(data string) (string, error) {
	if data == "" {
		return "", fmt.Errorf("multiple line data is required")
	}
	var payload multiplePayload
	dec := json.NewDecoder(strings.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&payload); err != nil {
		return "", fmt.Errorf("multiple line data invalid: %w", err)
	}
	if payload.Kind == "" {
		payload.Kind = "detailed"
	}
	if payload.Kind != "detailed" {
		return "", fmt.Errorf("multiple line data has invalid kind %q", payload.Kind)
	}
	for i, s := range payload.Sublines {
		if s.Name == "" {
			return "", fmt.Errorf("subline %d: name required", i)
		}
		if s.Quantity == "" {
			return "", fmt.Errorf("subline %d: quantity required", i)
		}
		if _, err := strconv.ParseFloat(s.Quantity, 64); err != nil {
			return "", fmt.Errorf("subline %d: quantity must be numeric", i)
		}
		if s.UnitPrice < 0 {
			return "", fmt.Errorf("subline %d: unit_price must be >= 0", i)
		}
		if payload.Sublines[i].Option == nil {
			defaultOption := false
			payload.Sublines[i].Option = &defaultOption
		}
	}

	clean, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(clean), nil
}
