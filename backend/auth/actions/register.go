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

	insertResp, err := s.userClient.CreateUser(ctx, &userGrpc.CreateUserRequest{
		Email: normalizedEmail,
	})
	if err != nil {
		return &authGrpc.FormGenericResponse{
			Success: false,
			Code:    CodeUserServiceError,
		}, err
	}
	if !insertResp.GetSuccess() {
		code := insertResp.GetCode()
		if code == 0 {
			code = CodeUserServiceError
		}
		return &authGrpc.FormGenericResponse{
			Success: false,
			Code:    code,
		}, nil
	}

	userID := insertResp.GetUserId()

	hashedPassword, err := services.HashPassword(req.Password)
	if err != nil {
		s.rollbackUser(ctx, userID)
		return &authGrpc.FormGenericResponse{
			Success: false,
			Code:    CodeInternalError,
		}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.rollbackUser(ctx, userID)
		return &authGrpc.FormGenericResponse{
			Success: false,
			Code:    CodeInternalError,
		}, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", int64(2026052901)); err != nil {
		s.rollbackUser(ctx, userID)
		return &authGrpc.FormGenericResponse{
			Success: false,
			Code:    CodeInternalError,
		}, err
	}

	role := "free_user"
	var authCount int64
	if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM auth").Scan(&authCount); err != nil {
		s.rollbackUser(ctx, userID)
		return &authGrpc.FormGenericResponse{
			Success: false,
			Code:    CodeInternalError,
		}, err
	}
	if authCount == 0 {
		role = "super_admin"
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO auth (user_id, email, password, role, account_status, subscription_tier) VALUES ($1, $2, $3, $4, $5, $6)",
		userID,
		normalizedEmail,
		hashedPassword,
		role,
		"active",
		"free",
	)
	if err != nil {
		s.rollbackUser(ctx, userID)
		return &authGrpc.FormGenericResponse{
			Success: false,
			Code:    CodeInternalError,
		}, err
	}

	if err = tx.Commit(); err != nil {
		s.rollbackUser(ctx, userID)
		return &authGrpc.FormGenericResponse{
			Success: false,
			Code:    CodeInternalError,
		}, err
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
