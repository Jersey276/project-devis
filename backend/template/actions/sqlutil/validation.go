package sqlutil

import templateGrpc "project-devis-template/services/grpc"

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

func ValidateTemplateType(t string) bool {
	return validTypes[t]
}

func Required(field string) *templateGrpc.ValidationError {
	return &templateGrpc.ValidationError{Field: field, Message: "Champ requis."}
}

func Invalid(field, message string) *templateGrpc.ValidationError {
	return &templateGrpc.ValidationError{Field: field, Message: message}
}
