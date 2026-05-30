package actions

import (
	"context"
	"database/sql"
	"project-devis-auth/services"
	authGrpc "project-devis-auth/services/grpc"
	"strings"
)

func (s *Server) Login(ctx context.Context, req *authGrpc.LoginRequest) (*authGrpc.LoginResponse, error) {
	emailInput := strings.ToLower(strings.TrimSpace(req.Email))

	var storedPassword, userID, email, role, accountStatus, subscriptionTier string
	var sessionVersion int32
	err := s.db.QueryRowContext(ctx,
		"SELECT email, password, user_id, role, account_status, subscription_tier, session_version FROM auth WHERE email = $1",
		emailInput,
	).Scan(&email, &storedPassword, &userID, &role, &accountStatus, &subscriptionTier, &sessionVersion)
	if err != nil {
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

	accessToken, err := services.GenerateAccessToken(email, userID, role, accountStatus, subscriptionTier, sessionVersion)
	if err != nil {
		code := CodeInternalError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, err
	}

	refreshToken, err := services.GenerateRefreshToken(ctx, s.db, userID, req.RememberMe)
	if err != nil {
		code := CodeInternalError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, err
	}

	return &authGrpc.LoginResponse{
		Success:      true,
		Token:        &accessToken,
		RefreshToken: &refreshToken,
	}, nil
}
