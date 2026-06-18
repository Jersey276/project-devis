package fee

import quoteGrpc "project-devis-quote/services/grpc"

const (
	CategoryFixed   = "fixed"
	CategoryService = "service"
)

// validateInput collects field errors common to Create and Update. fee_id and
// user_id are checked by the caller since their field names differ per RPC.
func validateInput(category, name string) []*quoteGrpc.ValidationError {
	var fieldErrors []*quoteGrpc.ValidationError
	if name == "" {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "name", Message: "Champ requis."})
	}
	if category == "" {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "category", Message: "Champ requis."})
	} else if category != CategoryFixed && category != CategoryService {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "category", Message: "Catégorie invalide."})
	}
	return fieldErrors
}
