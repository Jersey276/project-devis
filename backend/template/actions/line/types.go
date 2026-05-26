package line

import "encoding/json"

type TemplateLine struct {
	LineID     string          `json:"line_id"`
	TemplateID string          `json:"template_id"`
	Type       string          `json:"type"`
	Name       string          `json:"name"`
	Quantity   string          `json:"quantity"`
	Unit       *string         `json:"unit"`
	UnitPrice  int64           `json:"unit_price"`
	Data       json.RawMessage `json:"data"`
	Position   int             `json:"position"`
	TaxID      *int32          `json:"tax_id"`
	CreatedAt  string          `json:"created_at"`
	UpdatedAt  string          `json:"updated_at"`
}
