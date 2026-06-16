package actions

import (
	"context"
	"database/sql"
	"log"
	"net/mail"
	"project-devis-auth/services"
	authGrpc "project-devis-auth/services/grpc"
	userGrpc "project-devis-auth/services/user_auth"
	"strings"
)

func (s *Server) Register(ctx context.Context, req *authGrpc.RegisterRequest) (*authGrpc.FormGenericResponse, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))

	if fieldErrors := validateRegisterRequest(req); len(fieldErrors) > 0 {
		return &authGrpc.FormGenericResponse{
			Success:     false,
			Code:        CodeInternalError,
			FieldErrors: fieldErrors,
		}, nil
	}

	var existingEmail string
	err := s.db.QueryRowContext(ctx, "SELECT email FROM auth WHERE email = $1", normalizedEmail).Scan(&existingEmail)
	if err == nil {
		return &authGrpc.FormGenericResponse{
			Success: false,
			Code:    CodeUserAlreadyExists,
			FieldErrors: []*authGrpc.FormFieldError{
				{Field: "email", ErrorCode: []int32{FieldErrAlreadyInUse}},
			},
		}, nil
	}
	if err != sql.ErrNoRows {
		return &authGrpc.FormGenericResponse{
			Success: false,
			Code:    CodeInternalError,
		}, err
	}

	hashedPassword, err := services.HashPassword(req.Password)
	if err != nil {
		return &authGrpc.FormGenericResponse{
			Success: false,
			Code:    CodeInternalError,
		}, err
	}

	userID, err := s.provisionUser(ctx, normalizedEmail, hashedPassword, false)
	if err != nil {
		if provErr, ok := err.(*provisionError); ok {
			return &authGrpc.FormGenericResponse{
				Success: false,
				Code:    provErr.Code(),
			}, nil
		}
		return &authGrpc.FormGenericResponse{
			Success: false,
			Code:    CodeInternalError,
		}, err
	}

	if token, tokenErr := services.GenerateEmailVerificationToken(ctx, s.db, userID); tokenErr == nil {
		if sendErr := s.emailSender.SendEmailVerification(normalizedEmail, token); sendErr != nil {
			log.Printf("send verification email failed for user=%s: %v", userID, sendErr)
		}
	} else {
		log.Printf("generate verification token failed for user=%s: %v", userID, tokenErr)
	}

	return &authGrpc.FormGenericResponse{
		Success: true,
		Code:    CodeSuccess,
	}, nil
}

func validateRegisterRequest(req *authGrpc.RegisterRequest) []*authGrpc.FormFieldError {
	var fieldErrors []*authGrpc.FormFieldError

	if strings.TrimSpace(req.Email) == "" {
		fieldErrors = append(fieldErrors, &authGrpc.FormFieldError{
			Field:     "email",
			ErrorCode: []int32{FieldErrRequired},
		})
	} else if _, err := mail.ParseAddress(strings.TrimSpace(req.Email)); err != nil {
		fieldErrors = append(fieldErrors, &authGrpc.FormFieldError{
			Field:     "email",
			ErrorCode: []int32{FieldErrInvalidFormat},
		})
	}

	if req.Password == "" {
		fieldErrors = append(fieldErrors, &authGrpc.FormFieldError{
			Field:     "password",
			ErrorCode: []int32{FieldErrRequired},
		})
	} else if passwordCodes := passwordPolicyFieldErrors(req.Password); len(passwordCodes) > 0 {
		fieldErrors = append(fieldErrors, &authGrpc.FormFieldError{
			Field:     "password",
			ErrorCode: passwordCodes,
		})
	}

	return fieldErrors
}

func (s *Server) rollbackUser(ctx context.Context, userID string) {
	_, err := s.userClient.DeleteUser(ctx, &userGrpc.DeleteUserRequest{
		UserId: userID,
	})
	if err != nil {
		log.Printf("rollback: failed to delete user %s: %v", userID, err)
	}
}
