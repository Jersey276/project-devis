package actions

import (
	"net/mail"
	"strings"
	"unicode"

	authGrpc "project-devis-auth/services/grpc"
)

const minPasswordLength = 12

var allowedOAuthProviders = map[string]bool{
	"google":    true,
	"github":    true,
	"microsoft": true,
}

var validRoles = map[string]bool{
	"free_user":   true,
	"super_admin": true,
}

var validTiers = map[string]bool{
	"free":       true,
	"pro":        true,
	"enterprise": true,
}

func validEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func isStrongPassword(password string) bool {
	if len(password) < minPasswordLength {
		return false
	}
	return hasPasswordRequiredClasses(password)
}

func hasPasswordRequiredClasses(password string) bool {
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	return hasLower && hasUpper && hasDigit && hasSpecial
}

func passwordPolicyFieldErrors(password string) []int32 {
	var codes []int32
	if len(password) < minPasswordLength {
		codes = append(codes, FieldErrTooShort)
	}
	if !hasPasswordRequiredClasses(password) {
		codes = append(codes, FieldErrInvalidFormat)
	}
	return codes
}

func validateRegisterRequest(req *authGrpc.RegisterRequest) []*authGrpc.FormFieldError {
	var fieldErrors []*authGrpc.FormFieldError

	if strings.TrimSpace(req.Email) == "" {
		fieldErrors = append(fieldErrors, &authGrpc.FormFieldError{
			Field:     "email",
			ErrorCode: []int32{FieldErrRequired},
		})
	} else if _, err := mail.ParseAddress(strings.TrimSpace(req.Email)); err != nil {
		fieldErrors = append(fieldErrors, &authGrpc.FormFieldError{
			Field:     "email",
			ErrorCode: []int32{FieldErrInvalidFormat},
		})
	}

	if req.Password == "" {
		fieldErrors = append(fieldErrors, &authGrpc.FormFieldError{
			Field:     "password",
			ErrorCode: []int32{FieldErrRequired},
		})
	} else if passwordCodes := passwordPolicyFieldErrors(req.Password); len(passwordCodes) > 0 {
		fieldErrors = append(fieldErrors, &authGrpc.FormFieldError{
			Field:     "password",
			ErrorCode: passwordCodes,
		})
	}

	return fieldErrors
}

// validateOAuthIdentity normalizes and validates an OAuth identity payload
// shared by OAuthLogin and LinkOAuthIdentity. It returns the canonical
// provider, provider_user_id and email, plus a response code: CodeSuccess when
// valid, CodeInvalidInput for a bad provider/sub/email, or
// CodeOAuthEmailNotVerified when the provider did not attest the email.
func validateOAuthIdentity(rawProvider, rawProviderUserID, rawEmail string, emailVerified bool) (provider, providerUserID, email string, code int32) {
	provider = strings.ToLower(strings.TrimSpace(rawProvider))
	providerUserID = strings.TrimSpace(rawProviderUserID)
	email = strings.ToLower(strings.TrimSpace(rawEmail))

	if !allowedOAuthProviders[provider] || providerUserID == "" {
		return provider, providerUserID, email, CodeInvalidInput
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return provider, providerUserID, email, CodeInvalidInput
	}
	if !emailVerified {
		return provider, providerUserID, email, CodeOAuthEmailNotVerified
	}
	return provider, providerUserID, email, CodeSuccess
}
