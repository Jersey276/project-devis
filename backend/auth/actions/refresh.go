package actions

import (
	"context"

	"project-devis-auth/services"
	authGrpc "project-devis-auth/services/grpc"
	userGrpc "project-devis-auth/services/user_auth"
)

func (s *Server) RefreshToken(ctx context.Context, req *authGrpc.RefreshTokenRequest) (*authGrpc.LoginResponse, error) {
	userID, rememberMe, err := services.ValidateRefreshToken(ctx, s.db, req.RefreshToken)
	if err != nil {
		code := CodeInvalidRefreshToken
		return &authGrpc.LoginResponse{Success: false, Code: &code}, nil
	}

	accessInfo, err := s.userClient.GetUserAccessInfo(ctx, &userGrpc.GetUserAccessInfoRequest{UserId: userID})
	if err != nil {
		code := CodeUserServiceError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, err
	}
	if !accessInfo.GetSuccess() {
		_ = services.DeleteRefreshToken(ctx, s.db, req.RefreshToken)
		if accessInfo.GetCode() == userServiceCodeNotFound {
			code := CodeInvalidRefreshToken
			return &authGrpc.LoginResponse{Success: false, Code: &code}, nil
		}
		code := CodeUserServiceError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, nil
	}
	if accessInfo.GetSuspended() {
		_ = services.DeleteRefreshToken(ctx, s.db, req.RefreshToken)
		code := CodeInvalidRefreshToken
		return &authGrpc.LoginResponse{Success: false, Code: &code}, nil
	}

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

	// Rotate refresh token after successful validation.
	_ = services.DeleteRefreshToken(ctx, s.db, req.RefreshToken)

	accessToken, err := services.GenerateAccessToken(accessInfo.GetEmail(), userID, role, accountStatus, subscriptionTier, sessionVersion)
	if err != nil {
		code := CodeInternalError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, err
	}

	newRefreshToken, err := services.GenerateRefreshToken(ctx, s.db, userID, rememberMe)
	if err != nil {
		code := CodeInternalError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, err
	}

	code := CodeSuccess
	return &authGrpc.LoginResponse{
		Success:      true,
		Code:         &code,
		Token:        &accessToken,
		RefreshToken: &newRefreshToken,
		RememberMe:   &rememberMe,
	}, nil
}
