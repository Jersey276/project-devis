package actions

import (
	"context"
	"database/sql"
	"strings"

	authGrpc "project-devis-auth/services/grpc"
)

// LinkOAuthIdentity attaches a verified OAuth identity to an already
// authenticated account. Unlike OAuthLogin it never provisions a user and never
// issues tokens; the gateway resolves user_id from the caller's JWT. The OAuth
// email is trusted only when email_verified is true.
//
// Anti-account-takeover: if the identity is already linked to a *different*
// user, the request is refused with CodeOAuthIdentityTaken. Re-linking the same
// identity to the same user is idempotent.
func (s *Server) LinkOAuthIdentity(ctx context.Context, req *authGrpc.LinkOAuthIdentityRequest) (*authGrpc.GenericResponse, error) {
	userID := strings.TrimSpace(req.GetUserId())
	provider, providerUserID, normalizedEmail, code := validateOAuthIdentity(
		req.GetProvider(), req.GetProviderUserId(), req.GetEmail(), req.GetEmailVerified(),
	)
	if userID == "" || code != CodeSuccess {
		if code == CodeSuccess {
			code = CodeInvalidInput
		}
		return &authGrpc.GenericResponse{Success: false, Code: code}, nil
	}

	// Does this identity already exist?
	var ownerID string
	err := s.db.QueryRowContext(
		ctx,
		"SELECT user_id FROM oauth_identities WHERE provider = $1 AND provider_user_id = $2",
		provider, providerUserID,
	).Scan(&ownerID)
	switch {
	case err == nil:
		if ownerID == userID {
			return &authGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil // idempotent
		}
		return &authGrpc.GenericResponse{Success: false, Code: CodeOAuthIdentityTaken}, nil
	case err != sql.ErrNoRows:
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	if err := s.linkOAuthIdentity(ctx, provider, providerUserID, userID, normalizedEmail); err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}
	return &authGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

// UnlinkOAuthIdentity removes a linked provider. It refuses to remove the only
// remaining login method (an account with no password and a single identity)
// to avoid locking the user out.
func (s *Server) UnlinkOAuthIdentity(ctx context.Context, req *authGrpc.UnlinkOAuthIdentityRequest) (*authGrpc.GenericResponse, error) {
	userID := strings.TrimSpace(req.GetUserId())
	provider := strings.ToLower(strings.TrimSpace(req.GetProvider()))

	if userID == "" || !allowedOAuthProviders[provider] {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	hasPassword, identityCount, err := s.loginMethods(ctx, userID)
	if err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}
	if !hasPassword && identityCount <= 1 {
		return &authGrpc.GenericResponse{Success: false, Code: CodeLastLoginMethod}, nil
	}

	if _, err := s.db.ExecContext(
		ctx,
		"DELETE FROM oauth_identities WHERE user_id = $1 AND provider = $2",
		userID, provider,
	); err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}
	return &authGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

// ListOAuthIdentities returns the providers linked to a user plus whether the
// account also has a password (so the frontend can guard the last login method).
func (s *Server) ListOAuthIdentities(ctx context.Context, req *authGrpc.ListOAuthIdentitiesRequest) (*authGrpc.ListOAuthIdentitiesResponse, error) {
	userID := strings.TrimSpace(req.GetUserId())
	if userID == "" {
		return &authGrpc.ListOAuthIdentitiesResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	hasPassword, err := s.hasPassword(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &authGrpc.ListOAuthIdentitiesResponse{Success: false, Code: CodeUserNotFound}, nil
		}
		return &authGrpc.ListOAuthIdentitiesResponse{Success: false, Code: CodeInternalError}, err
	}

	rows, err := s.db.QueryContext(
		ctx,
		"SELECT provider, email FROM oauth_identities WHERE user_id = $1 ORDER BY provider",
		userID,
	)
	if err != nil {
		return &authGrpc.ListOAuthIdentitiesResponse{Success: false, Code: CodeInternalError}, err
	}
	defer rows.Close()

	var identities []*authGrpc.OAuthIdentity
	for rows.Next() {
		var provider, email string
		if err := rows.Scan(&provider, &email); err != nil {
			return &authGrpc.ListOAuthIdentitiesResponse{Success: false, Code: CodeInternalError}, err
		}
		identities = append(identities, &authGrpc.OAuthIdentity{Provider: provider, Email: email})
	}
	if err := rows.Err(); err != nil {
		return &authGrpc.ListOAuthIdentitiesResponse{Success: false, Code: CodeInternalError}, err
	}

	return &authGrpc.ListOAuthIdentitiesResponse{
		Success:     true,
		Code:        CodeSuccess,
		Identities:  identities,
		HasPassword: hasPassword,
	}, nil
}

// hasPassword reports whether the account has a password login. It returns
// sql.ErrNoRows when the user does not exist.
func (s *Server) hasPassword(ctx context.Context, userID string) (bool, error) {
	var hasPassword bool
	err := s.db.QueryRowContext(
		ctx,
		"SELECT password IS NOT NULL FROM auth WHERE user_id = $1",
		userID,
	).Scan(&hasPassword)
	return hasPassword, err
}

// loginMethods reports whether the account has a password and how many OAuth
// identities are linked, used to guard against removing the last login method.
// Both values come from a single round-trip.
func (s *Server) loginMethods(ctx context.Context, userID string) (hasPassword bool, identityCount int, err error) {
	err = s.db.QueryRowContext(
		ctx,
		`SELECT a.password IS NOT NULL,
		        (SELECT COUNT(*) FROM oauth_identities o WHERE o.user_id = a.user_id)
		 FROM auth a WHERE a.user_id = $1`,
		userID,
	).Scan(&hasPassword, &identityCount)
	return hasPassword, identityCount, err
}
