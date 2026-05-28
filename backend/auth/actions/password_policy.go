package actions

import "unicode"

const minPasswordLength = 12

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
