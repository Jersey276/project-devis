package actions

import (
	"context"
	"database/sql"
	"log"

	authGrpc "project-devis-auth/services/grpc"
	userGrpc "project-devis-auth/services/user_auth"
)

// OAuthLogin authenticates or provisions a user from a verified OAuth identity.
// Resolution order: (1) existing identity → login; (2) existing account with the
// same email → link the identity and login; (3) otherwise provision a new
// account. The OAuth email is trusted only when email_verified is true.
func (s *Server) OAuthLogin(ctx context.Context, req *authGrpc.OAuthLoginRequest) (*authGrpc.LoginResponse, error) {
	provider, providerUserID, normalizedEmail, code := validateOAuthIdentity(
		req.GetProvider(), req.GetProviderUserId(), req.GetEmail(), req.GetEmailVerified(),
	)
	if code != CodeSuccess {
		return &authGrpc.LoginResponse{Success: false, Code: &code}, nil
	}

	// (1) Existing identity → straight login.
	var userID string
	err := s.db.QueryRowContext(
		ctx,
		"SELECT user_id FROM oauth_identities WHERE provider = $1 AND provider_user_id = $2",
		provider, providerUserID,
	).Scan(&userID)
	if err == nil {
		s.touchLastLogin(ctx, userID)
		return s.issueLoginTokens(ctx, userID, normalizedEmail, req.GetRememberMe())
	}
	if err != sql.ErrNoRows {
		code := CodeInternalError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, err
	}

	// (2) Existing account with the same email → link identity and login.
	err = s.db.QueryRowContext(ctx, "SELECT user_id FROM auth WHERE email = $1", normalizedEmail).Scan(&userID)
	if err == nil {
		if err := s.linkOAuthIdentity(ctx, provider, providerUserID, userID, normalizedEmail); err != nil {
			code := CodeInternalError
			return &authGrpc.LoginResponse{Success: false, Code: &code}, err
		}
		// OAuth proves control of the inbox: mark the account verified.
		if _, err := s.db.ExecContext(ctx, "UPDATE auth SET email_verified = true WHERE user_id = $1", userID); err != nil {
			log.Printf("oauth: failed to mark email verified for user %s: %v", userID, err)
		}
		s.touchLastLogin(ctx, userID)
		return s.issueLoginTokens(ctx, userID, normalizedEmail, req.GetRememberMe())
	}
	if err != sql.ErrNoRows {
		code := CodeInternalError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, err
	}

	// (3) Provision a new account (NULL password, email already verified).
	userID, err = s.provisionUser(ctx, normalizedEmail, "", true)
	if err != nil {
		if provErr, ok := err.(*provisionError); ok {
			code := provErr.Code()
			return &authGrpc.LoginResponse{Success: false, Code: &code}, nil
		}
		code := CodeInternalError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, err
	}
	if err := s.linkOAuthIdentity(ctx, provider, providerUserID, userID, normalizedEmail); err != nil {
		code := CodeInternalError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, err
	}

	return s.issueLoginTokens(ctx, userID, normalizedEmail, req.GetRememberMe())
}

// linkOAuthIdentity inserts the identity, tolerating a concurrent duplicate
// callback via ON CONFLICT DO NOTHING (the UNIQUE(provider, provider_user_id)
// constraint makes the insert idempotent).
func (s *Server) linkOAuthIdentity(ctx context.Context, provider, providerUserID, userID, email string) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO oauth_identities (provider, provider_user_id, user_id, email)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (provider, provider_user_id) DO NOTHING`,
		provider, providerUserID, userID, email,
	)
	return err
}

// touchLastLogin updates the users-service last-login metadata, non-blocking.
func (s *Server) touchLastLogin(ctx context.Context, userID string) {
	if _, err := s.userClient.TouchUserLastLogin(ctx, &userGrpc.TouchUserLastLoginRequest{UserId: userID}); err != nil {
		log.Printf("touch last login failed for user %s: %v", userID, err)
	}
}
