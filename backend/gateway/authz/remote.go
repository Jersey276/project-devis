package authz

import (
	"context"
	"errors"
)

var ErrRemoteNotConfigured = errors.New("remote authorizer not configured")

type RemoteAuthorizer struct{}

func NewRemoteAuthorizer() *RemoteAuthorizer {
	return &RemoteAuthorizer{}
}

func (a *RemoteAuthorizer) Can(_ context.Context, _ Subject, _ Action, _ Resource) (Decision, error) {
	return Decision{Allowed: false, Reason: "AUTHZ_UNAVAILABLE"}, ErrRemoteNotConfigured
}
