package authz

import "context"

type LocalAuthorizer struct{}

func NewLocalAuthorizer() *LocalAuthorizer {
	return &LocalAuthorizer{}
}

func (a *LocalAuthorizer) Can(_ context.Context, subject Subject, action Action, resource Resource) (Decision, error) {
	if subject.AccountStatus == "suspended" && action != ActionRead {
		return Decision{Allowed: false, Reason: "ACCOUNT_SUSPENDED"}, nil
	}

	if resource == ResourceAdminCountries || resource == ResourceAdminCountryGroup || resource == ResourceAdminTaxes {
		if subject.AccountStatus != "active" {
			return Decision{Allowed: false, Reason: "ACCOUNT_INACTIVE"}, nil
		}
		if subject.Role != "super_admin" {
			return Decision{Allowed: false, Reason: "ADMIN_REQUIRED"}, nil
		}
	}

	return Decision{Allowed: true, Reason: "ALLOWED"}, nil
}
