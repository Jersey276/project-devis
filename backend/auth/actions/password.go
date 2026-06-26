package actions

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"project-devis-auth/services"
	authGrpc "project-devis-auth/services/grpc"
	userGrpc "project-devis-auth/services/user_auth"
)

func (s *Server) ResetPassword(ctx context.Context, req *authGrpc.ResetPasswordRequest) (*authGrpc.GenericResponse, error) {
	if !validEmail(req.Email) {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidCredentials}, nil
	}

	accessInfo, err := s.userClient.GetUserAccessInfoByEmail(ctx, &userGrpc.GetUserAccessInfoByEmailRequest{Email: req.Email})
	if err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeUserServiceError}, err
	}
	if !accessInfo.GetSuccess() {
		if accessInfo.GetCode() == userServiceCodeNotFound {
			// Anti-enumeration: keep a generic success when the account does not exist.
			return &authGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
		}
		return &authGrpc.GenericResponse{Success: false, Code: CodeUserServiceError}, nil
	}

	resetToken, err := services.GeneratePasswordResetToken(ctx, s.db, accessInfo.GetUserId())
	if err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	if err := s.emailSender.SendPasswordReset(accessInfo.GetEmail(), resetToken); err != nil {
		// Keep anti-enumeration behavior for callers.
		log.Printf("password reset email send failed for email=%s: %v", accessInfo.GetEmail(), err)
	}

	return &authGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

func (s *Server) UpdatePassword(ctx context.Context, req *authGrpc.UpdatePasswordRequest) (*authGrpc.GenericResponse, error) {
	if !validEmail(req.Email) {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidCredentials}, nil
	}
	if req.OldPassword == "" {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidCredentials}, nil
	}
	if !isStrongPassword(req.NewPassword) {
		return &authGrpc.GenericResponse{Success: false, Code: CodeWeakPassword}, nil
	}

	accessInfo, err := s.userClient.GetUserAccessInfoByEmail(ctx, &userGrpc.GetUserAccessInfoByEmailRequest{Email: req.Email})
	if err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeUserServiceError}, err
	}
	if !accessInfo.GetSuccess() {
		if accessInfo.GetCode() == userServiceCodeNotFound {
			return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidCredentials}, nil
		}
		return &authGrpc.GenericResponse{Success: false, Code: CodeUserServiceError}, nil
	}

	var storedPasswordHash string
	if err := s.db.QueryRowContext(ctx, "SELECT password FROM auth WHERE user_id = $1", accessInfo.GetUserId()).Scan(&storedPasswordHash); err != nil {
		if err == sql.ErrNoRows {
			return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidCredentials}, nil
		}
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	if !services.VerifyPassword(req.OldPassword, storedPasswordHash) {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidCredentials}, nil
	}

	hashedNewPassword, err := services.HashPassword(req.NewPassword)
	if err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "UPDATE auth SET password = $1, session_version = session_version + 1 WHERE user_id = $2", hashedNewPassword, accessInfo.GetUserId()); err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	if err := services.DeleteOtherRefreshTokensTx(ctx, tx, accessInfo.GetUserId(), req.CurrentRefreshToken); err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	if err := tx.Commit(); err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	return &authGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
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

	if _, err := tx.ExecContext(ctx, "UPDATE auth SET password = $1, session_version = session_version + 1 WHERE user_id = $2", hashedPassword, userID); err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE user_id = $1", userID); err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	if err := services.ConsumePasswordResetToken(ctx, tx, req.Token); err != nil {
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
