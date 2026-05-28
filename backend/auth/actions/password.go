package actions

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/mail"
	"project-devis-auth/services"
	authGrpc "project-devis-auth/services/grpc"
)

func (s *Server) ResetPassword(ctx context.Context, req *authGrpc.ResetPasswordRequest) (*authGrpc.GenericResponse, error) {
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return &authGrpc.GenericResponse{
			Success: false,
			Code:    CodeInvalidCredentials,
		}, nil
	}

	var userID string
	err := s.db.QueryRowContext(ctx, "SELECT user_id FROM auth WHERE email = $1", req.Email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Anti-enumeration: same success response when no account is found.
			return &authGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
		}
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	resetToken, err := services.GeneratePasswordResetToken(ctx, s.db, userID)
	if err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	if err := s.emailSender.SendPasswordReset(req.Email, resetToken); err != nil {
		// Keep anti-enumeration behavior: callers always receive success.
		log.Printf("password reset email send failed for email=%s: %v", req.Email, err)
	}

	return &authGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

func (s *Server) UpdatePassword(ctx context.Context, req *authGrpc.UpdatePasswordRequest) (*authGrpc.GenericResponse, error) {
	return &authGrpc.GenericResponse{
		Success: false,
		Code:    CodeNotImplemented,
	}, nil
}

func (s *Server) ConfirmResetPassword(ctx context.Context, req *authGrpc.ConfirmResetPasswordRequest) (*authGrpc.GenericResponse, error) {
	if req.Token == "" {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidResetToken}, nil
	}
	if !isStrongPassword(req.NewPassword) {
		return &authGrpc.GenericResponse{Success: false, Code: CodeWeakPassword}, nil
	}

	userID, err := services.ValidatePasswordResetToken(ctx, s.db, req.Token)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrPasswordResetTokenExpired):
			return &authGrpc.GenericResponse{Success: false, Code: CodeExpiredResetToken}, nil
		case errors.Is(err, services.ErrPasswordResetTokenUsed), errors.Is(err, services.ErrPasswordResetTokenNotFound):
			return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidResetToken}, nil
		default:
			return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
		}
	}

	hashedPassword, err := services.HashPassword(req.NewPassword)
	if err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "UPDATE auth SET password = $1 WHERE user_id = $2", hashedPassword, userID); err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE user_id = $1", userID); err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	if err := services.ConsumePasswordResetTokenTx(ctx, tx, req.Token); err != nil {
		switch {
		case errors.Is(err, services.ErrPasswordResetTokenNotFound):
			return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidResetToken}, nil
		default:
			return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	return &authGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}
