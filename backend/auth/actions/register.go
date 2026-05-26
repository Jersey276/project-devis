package actions

import (
	"context"
	"database/sql"
	"log"
	"net/mail"
	"project-devis-auth/services"
	authGrpc "project-devis-auth/services/grpc"
	userGrpc "project-devis-auth/services/user_auth"
)

func (s *Server) Register(ctx context.Context, req *authGrpc.RegisterRequest) (*authGrpc.FormGenericResponse, error) {
	if fieldErrors := validateRegisterRequest(req); len(fieldErrors) > 0 {
		return &authGrpc.FormGenericResponse{
			Success:     false,
			Code:        CodeInternalError,
			FieldErrors: fieldErrors,
		}, nil
	}

	var existingEmail string
	err := s.db.QueryRowContext(ctx, "SELECT email FROM auth WHERE email = $1", req.Email).Scan(&existingEmail)
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
		Email: req.Email,
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

	_, err = s.db.ExecContext(ctx, "INSERT INTO auth (user_id, email, password) VALUES ($1, $2, $3)", userID, req.Email, hashedPassword)
	if err != nil {
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

	if req.Email == "" {
		fieldErrors = append(fieldErrors, &authGrpc.FormFieldError{
			Field:     "email",
			ErrorCode: []int32{FieldErrRequired},
		})
	} else if _, err := mail.ParseAddress(req.Email); err != nil {
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
	} else if len(req.Password) < 8 {
		fieldErrors = append(fieldErrors, &authGrpc.FormFieldError{
			Field:     "password",
			ErrorCode: []int32{FieldErrTooShort},
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
