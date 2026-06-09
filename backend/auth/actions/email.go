package actions

import (
	"context"
	"errors"
	"log"

	"project-devis-auth/services"
	authGrpc "project-devis-auth/services/grpc"
)

func (s *Server) ResendEmailVerification(ctx context.Context, req *authGrpc.ResendEmailVerificationRequest) (*authGrpc.GenericResponse, error) {
	if req.UserId == "" {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	var email string
	var verified bool
	err := s.db.QueryRowContext(ctx,
		"SELECT email, email_verified FROM auth WHERE user_id = $1",
		req.UserId,
	).Scan(&email, &verified)
	if err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeUserNotFound}, nil
	}

	if verified {
		return &authGrpc.GenericResponse{Success: false, Code: CodeAlreadyVerified}, nil
	}

	token, err := services.GenerateEmailVerificationToken(ctx, s.db, req.UserId)
	if err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	if err := s.emailSender.SendEmailVerification(email, token); err != nil {
		log.Printf("resend verification email failed for user=%s: %v", req.UserId, err)
	}

	return &authGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

func (s *Server) VerifyEmail(ctx context.Context, req *authGrpc.VerifyEmailRequest) (*authGrpc.GenericResponse, error) {
	if req.Token == "" {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidVerificationToken}, nil
	}

	userID, err := services.ValidateEmailVerificationToken(ctx, s.db, req.Token)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmailVerificationTokenExpired):
			return &authGrpc.GenericResponse{Success: false, Code: CodeExpiredVerificationToken}, nil
		case errors.Is(err, services.ErrEmailVerificationTokenUsed),
			errors.Is(err, services.ErrEmailVerificationTokenNotFound):
			return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidVerificationToken}, nil
		default:
			return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "UPDATE auth SET email_verified=true WHERE user_id=$1", userID); err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	if err := services.ConsumeEmailVerificationToken(ctx, s.db, req.Token); err != nil {
		if !errors.Is(err, services.ErrEmailVerificationTokenNotFound) {
			log.Printf("consume email verification token failed for user=%s: %v", userID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	return &authGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}
