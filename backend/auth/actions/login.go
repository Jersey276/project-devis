package actions

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"project-devis-auth/services"
	authGrpc "project-devis-auth/services/grpc"
	userGrpc "project-devis-auth/services/user_auth"
)

const userServiceCodeNotFound int32 = 1001

func (s *Server) Login(ctx context.Context, req *authGrpc.LoginRequest) (*authGrpc.LoginResponse, error) {
	accessInfo, err := s.userClient.GetUserAccessInfoByEmail(ctx, &userGrpc.GetUserAccessInfoByEmailRequest{Email: strings.TrimSpace(req.Email)})
	if err != nil {
		code := CodeUserServiceError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, err
	}
	if !accessInfo.GetSuccess() {
		if accessInfo.GetCode() == userServiceCodeNotFound {
			code := CodeUserNotFound
			return &authGrpc.LoginResponse{Success: false, Code: &code}, nil
		}
		code := CodeUserServiceError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, nil
	}
	if accessInfo.GetSuspended() {
		code := CodeInvalidCredentials
		return &authGrpc.LoginResponse{Success: false, Code: &code}, nil
	}

	var storedPassword, role, accountStatus, subscriptionTier string
	var sessionVersion int32
	if err := s.db.QueryRowContext(
		ctx,
		"SELECT password, role, account_status, subscription_tier, session_version FROM auth WHERE user_id = $1",
		accessInfo.GetUserId(),
	).Scan(&storedPassword, &role, &accountStatus, &subscriptionTier, &sessionVersion); err != nil {
		if err == sql.ErrNoRows {
			code := CodeUserNotFound
			return &authGrpc.LoginResponse{Success: false, Code: &code}, nil
		}
		code := CodeInternalError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, err
	}

	if !services.VerifyPassword(req.Password, storedPassword) {
		code := CodeInvalidCredentials
		return &authGrpc.LoginResponse{Success: false, Code: &code}, nil
	}

	accessToken, err := services.GenerateAccessToken(accessInfo.GetEmail(), accessInfo.GetUserId(), role, accountStatus, subscriptionTier, sessionVersion)
	if err != nil {
		code := CodeInternalError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, err
	}

	refreshToken, err := services.GenerateRefreshToken(ctx, s.db, accessInfo.GetUserId(), req.RememberMe)
	if err != nil {
		code := CodeInternalError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, err
	}

	if _, err := s.userClient.TouchUserLastLogin(ctx, &userGrpc.TouchUserLastLoginRequest{UserId: accessInfo.GetUserId()}); err != nil {
		// Non-blocking metadata update.
		log.Printf("touch last login failed for user %s: %v", accessInfo.GetUserId(), err)
	}

	code := CodeSuccess
	return &authGrpc.LoginResponse{
		Success:      true,
		Code:         &code,
		Token:        &accessToken,
		RefreshToken: &refreshToken,
		RememberMe:   &req.RememberMe,
	}, nil
}
