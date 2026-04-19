package actions

import (
	"context"
	authGrpc "project-devis-auth/services/grpc"
)

func (s *Server) VerifyEmail(ctx context.Context, req *authGrpc.VerifyEmailRequest) (*authGrpc.GenericResponse, error) {
	return &authGrpc.GenericResponse{
		Success: false,
		Code:    CodeNotImplemented,
	}, nil
}
