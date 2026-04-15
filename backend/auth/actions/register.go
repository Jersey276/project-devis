package actions

import (
	"context"
	"database/sql"
	"log"
	"project-devis-auth/services"
	authGrpc "project-devis-auth/services/grpc"
	userGrpc "project-devis-auth/services/user_auth"
)

func (s *Server) Register(ctx context.Context, req *authGrpc.RegisterRequest) (*authGrpc.RegisterResponse, error) {
	var existingEmail string
	err := s.db.QueryRowContext(ctx, "SELECT email FROM auth WHERE email = $1", req.Email).Scan(&existingEmail)
	if err == nil {
		return &authGrpc.RegisterResponse{
			Success: false,
			Code:    CodeUserAlreadyExists,
		}, nil
	}
	if err != sql.ErrNoRows {
		return &authGrpc.RegisterResponse{
			Success: false,
			Code:    CodeInternalError,
		}, err
	}

	insertResp, err := s.userClient.InsertUser(ctx, &userGrpc.InsertUserRequest{
		Email:    req.Email,
		Username: req.Name,
	})
	if err != nil {
		return &authGrpc.RegisterResponse{
			Success: false,
			Code:    CodeUserServiceError,
		}, err
	}
	if !insertResp.GetSuccess() {
		code := insertResp.GetCode()
		if code == 0 {
			code = CodeUserServiceError
		}
		return &authGrpc.RegisterResponse{
			Success: false,
			Code:    code,
		}, nil
	}

	userID := insertResp.GetUserId()

	hashedPassword, err := services.HashPassword(req.Password)
	if err != nil {
		s.rollbackUser(ctx, userID)
		return &authGrpc.RegisterResponse{
			Success: false,
			Code:    CodeInternalError,
		}, err
	}

	_, err = s.db.ExecContext(ctx, "INSERT INTO auth (user_id, email, password) VALUES ($1, $2, $3)", userID, req.Email, hashedPassword)
	if err != nil {
		s.rollbackUser(ctx, userID)
		return &authGrpc.RegisterResponse{
			Success: false,
			Code:    CodeInternalError,
		}, err
	}

	return &authGrpc.RegisterResponse{
		Success: true,
		Code:    CodeSuccess,
	}, nil
}

func (s *Server) rollbackUser(ctx context.Context, userID string) {
	_, err := s.userClient.DeleteUser(ctx, &userGrpc.DeleteUserRequest{
		UserId: userID,
	})
	if err != nil {
		log.Printf("rollback: failed to delete user %s: %v", userID, err)
	}
}