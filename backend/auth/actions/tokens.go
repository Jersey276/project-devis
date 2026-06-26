package actions

import (
	"context"

	"project-devis-auth/services"
	authGrpc "project-devis-auth/services/grpc"
)

// issueLoginTokens loads the access-control fields for userID, generates an
// access token + refresh token, and returns a populated success LoginResponse.
// It centralizes the SELECT role/account_status/subscription_tier/session_version
// + GenerateAccessToken + GenerateRefreshToken sequence shared by Login,
// RefreshToken, and OAuthLogin.
func (s *Server) issueLoginTokens(ctx context.Context, userID, email string, rememberMe bool) (*authGrpc.LoginResponse, error) {
	var role, accountStatus, subscriptionTier string
	var sessionVersion int32
	if err := s.db.QueryRowContext(
		ctx,
		"SELECT role, account_status, subscription_tier, session_version FROM auth WHERE user_id = $1",
		userID,
	).Scan(&role, &accountStatus, &subscriptionTier, &sessionVersion); err != nil {
		code := CodeInternalError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, err
	}

	accessToken, err := services.GenerateAccessToken(email, userID, role, accountStatus, subscriptionTier, sessionVersion)
	if err != nil {
		code := CodeInternalError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, err
	}

	refreshToken, err := services.GenerateRefreshToken(ctx, s.db, userID, rememberMe)
	if err != nil {
		code := CodeInternalError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, err
	}

	code := CodeSuccess
	return &authGrpc.LoginResponse{
		Success:      true,
		Code:         &code,
		Token:        &accessToken,
		RefreshToken: &refreshToken,
		RememberMe:   &rememberMe,
	}, nil
}
