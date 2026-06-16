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

	// Rotate refresh token after successful validation; issueLoginTokens then
	// mints the new access + refresh pair.
	_ = services.DeleteRefreshToken(ctx, s.db, req.RefreshToken)

	return s.issueLoginTokens(ctx, userID, accessInfo.GetEmail(), rememberMe)
}
