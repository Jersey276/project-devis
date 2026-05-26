package actions

import (
	"context"
	authGrpc "project-devis-auth/services/grpc"
)

func (s *Server) ResetPassword(ctx context.Context, req *authGrpc.ResetPasswordRequest) (*authGrpc.GenericResponse, error) {
	return &authGrpc.GenericResponse{
		Success: false,
		Code:    CodeNotImplemented,
	}, nil
}

func (s *Server) UpdatePassword(ctx context.Context, req *authGrpc.UpdatePasswordRequest) (*authGrpc.GenericResponse, error) {
	return &authGrpc.GenericResponse{
		Success: false,
		Code:    CodeNotImplemented,
	}, nil
}
