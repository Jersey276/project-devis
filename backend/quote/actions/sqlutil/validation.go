package sqlutil

import quoteGrpc "project-devis-quote/services/grpc"

const (
	CategoryFixed   = "fixed"
	CategoryService = "service"
	TypeSimple      = "simple"
	TypeMultiple    = "multiple"
)

func Required(field string) *quoteGrpc.ValidationError {
	return &quoteGrpc.ValidationError{Field: field, Message: "Champ requis."}
}

func Invalid(field, message string) *quoteGrpc.ValidationError {
	return &quoteGrpc.ValidationError{Field: field, Message: message}
}

func NonNegative(field string) *quoteGrpc.ValidationError {
	return &quoteGrpc.ValidationError{Field: field, Message: "Doit être positif ou nul."}
}

func ValidateFeeInput(category, name string) []*quoteGrpc.ValidationError {
	var fieldErrors []*quoteGrpc.ValidationError
	if name == "" {
		fieldErrors = append(fieldErrors, Required("name"))
	}
	if category == "" {
		fieldErrors = append(fieldErrors, Required("category"))
	} else if category != CategoryFixed && category != CategoryService {
		fieldErrors = append(fieldErrors, Invalid("category", "Catégorie invalide."))
	}
	return fieldErrors
}
