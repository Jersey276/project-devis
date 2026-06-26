package actions

import (
	"context"
	"errors"
	"log"
	"strings"

	"project-devis-auth/services"
	authGrpc "project-devis-auth/services/grpc"
	userGrpc "project-devis-auth/services/user_auth"
)

func (s *Server) RequestEmailChange(ctx context.Context, req *authGrpc.RequestEmailChangeRequest) (*authGrpc.GenericResponse, error) {
	if req.UserId == "" {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}
	newEmail := strings.ToLower(strings.TrimSpace(req.NewEmail))
	if !validEmail(newEmail) {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	// Check the new email isn't already taken in the auth table.
	var existing string
	err := s.db.QueryRowContext(ctx, "SELECT user_id FROM auth WHERE email = $1 AND user_id != $2", newEmail, req.UserId).Scan(&existing)
	if err == nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeEmailAlreadyInUse}, nil
	}

	// Get current email for the sending address.
	var currentEmail string
	if err := s.db.QueryRowContext(ctx, "SELECT email FROM auth WHERE user_id = $1", req.UserId).Scan(&currentEmail); err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeUserNotFound}, nil
	}

	token, err := services.GenerateEmailChangeToken(ctx, s.db, req.UserId, newEmail)
	if err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	if err := s.emailSender.SendEmailChange(newEmail, token); err != nil {
		log.Printf("send email change failed for user=%s: %v", req.UserId, err)
	}

	return &authGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

func (s *Server) ConfirmEmailChange(ctx context.Context, req *authGrpc.ConfirmEmailChangeRequest) (*authGrpc.GenericResponse, error) {
	if req.Token == "" {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidEmailChangeToken}, nil
	}

	userID, newEmail, err := services.ValidateEmailChangeToken(ctx, s.db, req.Token)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmailChangeTokenExpired):
			return &authGrpc.GenericResponse{Success: false, Code: CodeExpiredEmailChangeToken}, nil
		case errors.Is(err, services.ErrEmailChangeTokenUsed),
			errors.Is(err, services.ErrEmailChangeTokenNotFound):
			return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidEmailChangeToken}, nil
		default:
			return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
		}
	}

	// Update the users service first (uniqueness guard lives there too).
	resp, err := s.userClient.UpdateUserEmail(ctx, &userGrpc.UpdateUserEmailRequest{
		UserId:   userID,
		NewEmail: newEmail,
	})
	if err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeUserServiceError}, err
	}
	if !resp.GetSuccess() {
		if resp.GetCode() == 1002 { // AlreadyExists
			return &authGrpc.GenericResponse{Success: false, Code: CodeEmailAlreadyInUse}, nil
		}
		return &authGrpc.GenericResponse{Success: false, Code: CodeUserServiceError}, nil
	}

	// Update auth table and bump session_version to invalidate existing tokens.
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		"UPDATE auth SET email = $1, session_version = session_version + 1 WHERE user_id = $2",
		newEmail, userID,
	); err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	if err := services.ConsumeEmailChangeToken(ctx, s.db, req.Token); err != nil {
		if !errors.Is(err, services.ErrEmailChangeTokenNotFound) {
			log.Printf("consume email change token failed for user=%s: %v", userID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	return &authGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}
