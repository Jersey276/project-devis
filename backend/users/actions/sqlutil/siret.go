package sqlutil

import "strings"

// NormalizeSIRET strips spaces from a SIRET/SIREN (users enter them grouped).
func NormalizeSIRET(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), " ", "")
}

// ValidateSIRET checks a recipient routing SIRET. Empty is allowed (optional).
// When set it must be 14 digits; when a SIREN is also set, the SIRET must start
// with it (a SIRET is its SIREN + a 5-digit establishment NIC). Returns "" if
// valid, else a user-facing French message for the `siret` field.
func ValidateSIRET(siret, siren string) string {
	siret = NormalizeSIRET(siret)
	if siret == "" {
		return ""
	}
	if len(siret) != 14 || !isDigits(siret) {
		return "Le SIRET doit comporter 14 chiffres."
	}
	if siren = NormalizeSIRET(siren); siren != "" && !strings.HasPrefix(siret, siren) {
		return "Le SIRET doit commencer par le SIREN."
	}
	return ""
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
