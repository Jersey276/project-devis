package template

import "encoding/json"

const (
	TypeQuoteDocument  = "quote_document"
	TypeQuoteLine      = "quote_line"
	TypeDocumentDesign = "document_design"
)

var validTypes = map[string]bool{
	TypeQuoteDocument:  true,
	TypeQuoteLine:      true,
	TypeDocumentDesign: true,
}

type Template struct {
	TemplateID     string          `json:"template_id"`
	UserID         string          `json:"user_id"`
	TemplateType   string          `json:"template_type"`
	TargetResource string          `json:"target_resource"`
	Name           string          `json:"name"`
	ArchivedAt     *string         `json:"archived_at"`
	PayloadVersion int             `json:"payload_version"`
	Payload        json.RawMessage `json:"payload"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
}
