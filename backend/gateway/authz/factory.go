package authz

import "os"

func NewFromEnv() Authorizer {
	if os.Getenv("AUTHZ_PROVIDER") == "remote" && os.Getenv("AUTHZ_REMOTE_STRICT") == "true" {
		return NewRemoteAuthorizer()
	}
	return NewLocalAuthorizer()
}
