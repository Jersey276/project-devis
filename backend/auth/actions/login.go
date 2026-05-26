package actions

import (
	"context"
	"database/sql"
	"project-devis-auth/services"
	authGrpc "project-devis-auth/services/grpc"
)

func (s *Server) Login(ctx context.Context, req *authGrpc.LoginRequest) (*authGrpc.LoginResponse, error) {
	var storedPassword, userID, email string
	err := s.db.QueryRowContext(ctx, "SELECT email, password, user_id FROM auth WHERE email = $1", req.Email).Scan(&email, &storedPassword, &userID)
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

	accessToken, err := services.GenerateAccessToken(email, userID)
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
