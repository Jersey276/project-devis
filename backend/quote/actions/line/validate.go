package line

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const (
	TypeSimple   = "simple"
	TypeMultiple = "multiple"
)

// ValidateData ensures the JSON payload matches the schema for the given line type.
// Returns the canonical JSON (defaulting Simple to "{}") if valid.
func ValidateData(lineType, data string) (string, error) {
	switch lineType {
	case TypeSimple:
		return validateSimple(data)
	case TypeMultiple:
		return validateMultiple(data)
	default:
		return "", fmt.Errorf("unknown line type %q", lineType)
	}
}

func validateSimple(data string) (string, error) {
	if data == "" || data == "null" {
		return "{}", nil
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return "", fmt.Errorf("simple line data must be a JSON object: %w", err)
	}
	if len(payload) != 0 {
		return "", fmt.Errorf("simple line data must be empty")
	}
	return "{}", nil
}

type subline struct {
	Name      string  `json:"name"`
	Quantity  string  `json:"quantity"`
	Unit      *string `json:"unit,omitempty"`
	UnitPrice int64   `json:"unit_price"`
}

type multiplePayload struct {
	Sublines []subline `json:"sublines"`
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
	if len(payload.Sublines) == 0 {
		return "", fmt.Errorf("multiple line must have at least one subline")
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
	}

	clean, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(clean), nil
}
