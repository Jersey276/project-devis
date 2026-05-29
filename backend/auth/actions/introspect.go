package actions

import (
	"context"
	"database/sql"
	"strings"

	"project-devis-auth/services"
	authGrpc "project-devis-auth/services/grpc"
)

func (s *Server) IntrospectToken(ctx context.Context, req *authGrpc.IntrospectTokenRequest) (*authGrpc.IntrospectTokenResponse, error) {
	token := strings.TrimSpace(req.GetToken())
	if token == "" {
		return &authGrpc.IntrospectTokenResponse{Success: false, Code: CodeInvalidCredentials}, nil
	}

	claims, err := services.ValidateAccessToken(token)
	if err != nil {
		return &authGrpc.IntrospectTokenResponse{Success: false, Code: CodeInvalidCredentials}, nil
	}

	var email, role, accountStatus, subscriptionTier string
	var sessionVersion int32
	err = s.db.QueryRowContext(ctx,
		"SELECT email, role, account_status, subscription_tier, session_version FROM auth WHERE user_id = $1",
		claims.UserID,
	).Scan(&email, &role, &accountStatus, &subscriptionTier, &sessionVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			return &authGrpc.IntrospectTokenResponse{Success: false, Code: CodeUserNotFound}, nil
		}
		return &authGrpc.IntrospectTokenResponse{Success: false, Code: CodeInternalError}, err
	}

	if claims.SessionVersion != sessionVersion {
		return &authGrpc.IntrospectTokenResponse{Success: false, Code: CodeSessionInvalidated}, nil
	}

	return &authGrpc.IntrospectTokenResponse{
		Success: true,
		Code:    CodeSuccess,
		Context: &authGrpc.AccessContext{
			UserId:           claims.UserID,
			Email:            email,
			Role:             role,
			AccountStatus:    accountStatus,
			SubscriptionTier: subscriptionTier,
			SessionVersion:   sessionVersion,
		},
	}, nil
}
