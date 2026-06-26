package actions

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"

	"project-devis-auth/services"
	authGrpc "project-devis-auth/services/grpc"
	userGrpc "project-devis-auth/services/user_auth"
)

func (s *Server) SendClientInvitation(ctx context.Context, req *authGrpc.SendClientInvitationRequest) (*authGrpc.GenericResponse, error) {
	if req.ProviderUserId == "" || req.ClientId == "" || req.ClientEmail == "" {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	rawToken, err := services.GenerateClientInvitationToken(ctx, s.db, req.ClientId, req.ProviderUserId)
	if err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	if err := s.emailSender.SendClientInvitation(req.ClientEmail, req.ClientName, rawToken); err != nil {
		// Anti-enumeration: log but do not expose email delivery failures to the caller.
		log.Printf("client invitation email failed: client_id=%s to=%s err=%v", req.ClientId, req.ClientEmail, err)
	}

	return &authGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

func (s *Server) AcceptClientInvitationNew(ctx context.Context, req *authGrpc.AcceptClientInvitationNewRequest) (*authGrpc.AcceptClientInvitationResponse, error) {
	if req.Token == "" {
		return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeInvalidInvitationToken}, nil
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))
	if !validEmail(normalizedEmail) {
		return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeInvalidCredentials}, nil
	}
	if !isStrongPassword(req.Password) {
		return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeWeakPassword}, nil
	}

	clientID, providerID, err := services.ValidateClientInvitationToken(ctx, s.db, req.Token)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrClientInvitationTokenExpired):
			return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeExpiredInvitationToken}, nil
		case errors.Is(err, services.ErrClientInvitationTokenUsed), errors.Is(err, services.ErrClientInvitationTokenNotFound):
			return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeInvalidInvitationToken}, nil
		default:
			return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeInternalError}, err
		}
	}

	// Check for duplicate email before provisioning.
	var existingEmail string
	dbErr := s.db.QueryRowContext(ctx, "SELECT email FROM auth WHERE email = $1", normalizedEmail).Scan(&existingEmail)
	if dbErr == nil {
		return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeUserAlreadyExists}, nil
	}
	if dbErr != sql.ErrNoRows {
		return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeInternalError}, dbErr
	}

	hashedPassword, err := services.HashPassword(req.Password)
	if err != nil {
		return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeInternalError}, err
	}

	// The invitation link itself confirms email ownership — provision with email_verified=true.
	newUserID, err := s.provisionUser(ctx, normalizedEmail, hashedPassword, true)
	if err != nil {
		if provErr, ok := err.(*provisionError); ok {
			return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: provErr.Code()}, nil
		}
		return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeInternalError}, err
	}

	linkResp, err := s.userClient.LinkClientUser(ctx, &userGrpc.LinkClientUserRequest{
		ClientId:     clientID,
		ProviderId:   providerID,
		LinkedUserId: newUserID,
	})
	if err != nil {
		s.rollbackUser(ctx, newUserID)
		return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeUserServiceError}, err
	}
	if !linkResp.GetSuccess() {
		s.rollbackUser(ctx, newUserID)
		if linkResp.GetCode() == 1004 { // users.AlreadyLinked
			return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeClientAlreadyLinked}, nil
		}
		return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeUserServiceError}, nil
	}

	// Consume token only after successful linking — allows retry on transient failure.
	if err := services.ConsumeClientInvitationToken(ctx, s.db, req.Token); err != nil {
		log.Printf("consume invitation token failed (link succeeded): client_id=%s err=%v", clientID, err)
	}

	loginResp, err := s.issueLoginTokens(ctx, newUserID, normalizedEmail, false)
	if err != nil {
		return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeInternalError}, err
	}

	return &authGrpc.AcceptClientInvitationResponse{
		Success:      true,
		Code:         CodeSuccess,
		Token:        loginResp.GetToken(),
		RefreshToken: loginResp.GetRefreshToken(),
		IsNewAccount: true,
	}, nil
}

func (s *Server) AcceptClientInvitationLinked(ctx context.Context, req *authGrpc.AcceptClientInvitationLinkedRequest) (*authGrpc.AcceptClientInvitationResponse, error) {
	if req.Token == "" || req.UserId == "" {
		return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeInvalidInvitationToken}, nil
	}

	clientID, providerID, err := services.ValidateClientInvitationToken(ctx, s.db, req.Token)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrClientInvitationTokenExpired):
			return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeExpiredInvitationToken}, nil
		case errors.Is(err, services.ErrClientInvitationTokenUsed), errors.Is(err, services.ErrClientInvitationTokenNotFound):
			return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeInvalidInvitationToken}, nil
		default:
			return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeInternalError}, err
		}
	}

	linkResp, err := s.userClient.LinkClientUser(ctx, &userGrpc.LinkClientUserRequest{
		ClientId:     clientID,
		ProviderId:   providerID,
		LinkedUserId: req.UserId,
	})
	if err != nil {
		return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeUserServiceError}, err
	}
	if !linkResp.GetSuccess() {
		if linkResp.GetCode() == 1004 { // users.AlreadyLinked
			return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeClientAlreadyLinked}, nil
		}
		return &authGrpc.AcceptClientInvitationResponse{Success: false, Code: CodeUserServiceError}, nil
	}

	if err := services.ConsumeClientInvitationToken(ctx, s.db, req.Token); err != nil {
		log.Printf("consume invitation token failed (link succeeded): client_id=%s err=%v", clientID, err)
	}

	return &authGrpc.AcceptClientInvitationResponse{
		Success:      true,
		Code:         CodeSuccess,
		IsNewAccount: false,
	}, nil
}
