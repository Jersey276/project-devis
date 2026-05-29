package authz

import "os"

func NewFromEnv() Authorizer {
	if os.Getenv("AUTHZ_PROVIDER") == "remote" {
		return NewRemoteAuthorizer()
	}
	return NewLocalAuthorizer()
}
