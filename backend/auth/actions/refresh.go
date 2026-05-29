package actions

import (
	"context"
	"project-devis-auth/services"
	authGrpc "project-devis-auth/services/grpc"
)

func (s *Server) RefreshToken(ctx context.Context, req *authGrpc.RefreshTokenRequest) (*authGrpc.LoginResponse, error) {
	userID, rememberMe, err := services.ValidateRefreshToken(ctx, s.db, req.RefreshToken)
	if err != nil {
		code := CodeInvalidRefreshToken
		return &authGrpc.LoginResponse{Success: false, Code: &code}, nil
	}

	// Get auth identity + access control fields for access-token claims.
	var email, role, accountStatus, subscriptionTier string
	var sessionVersion int32
	err = s.db.QueryRowContext(ctx,
		"SELECT email, role, account_status, subscription_tier, session_version FROM auth WHERE user_id = $1",
		userID,
	).Scan(&email, &role, &accountStatus, &subscriptionTier, &sessionVersion)
	if err != nil {
		code := CodeUserNotFound
		return &authGrpc.LoginResponse{Success: false, Code: &code}, nil
	}

	// Delete old refresh token (rotation)
	_ = services.DeleteRefreshToken(ctx, s.db, req.RefreshToken)

	accessToken, err := services.GenerateAccessToken(email, userID, role, accountStatus, subscriptionTier, sessionVersion)
	if err != nil {
		code := CodeInternalError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, err
	}

	// Preserve the original remember_me choice so cookie TTL stays consistent across rotations.
	newRefreshToken, err := services.GenerateRefreshToken(ctx, s.db, userID, rememberMe)
	if err != nil {
		code := CodeInternalError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, err
	}

	return &authGrpc.LoginResponse{
		Success:      true,
		Token:        &accessToken,
		RefreshToken: &newRefreshToken,
		RememberMe:   &rememberMe,
	}, nil
}
