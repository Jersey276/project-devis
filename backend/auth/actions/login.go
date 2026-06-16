package actions

import (
	"context"
	"database/sql"
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

	// password is nullable: OAuth-only accounts have no password. A NULL/empty
	// stored password can never match, so such accounts fail password login
	// cleanly with invalid credentials.
	var storedPassword sql.NullString
	if err := s.db.QueryRowContext(
		ctx,
		"SELECT password FROM auth WHERE user_id = $1",
		accessInfo.GetUserId(),
	).Scan(&storedPassword); err != nil {
		if err == sql.ErrNoRows {
			code := CodeUserNotFound
			return &authGrpc.LoginResponse{Success: false, Code: &code}, nil
		}
		code := CodeInternalError
		return &authGrpc.LoginResponse{Success: false, Code: &code}, err
	}

	if !storedPassword.Valid || !services.VerifyPassword(req.Password, storedPassword.String) {
		code := CodeInvalidCredentials
		return &authGrpc.LoginResponse{Success: false, Code: &code}, nil
	}

	resp, err := s.issueLoginTokens(ctx, accessInfo.GetUserId(), accessInfo.GetEmail(), req.RememberMe)
	if err != nil {
		return resp, err
	}

	s.touchLastLogin(ctx, accessInfo.GetUserId())

	return resp, nil
}
