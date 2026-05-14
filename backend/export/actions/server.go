package actions

import (
	"project-devis-export/quote"
	"project-devis-export/services/gotenberg"
	exportGrpc "project-devis-export/services/grpc"
	"project-devis-export/users"
)

type Server struct {
	exportGrpc.UnimplementedExportServiceServer
	quote     quote.QuoteServiceClient
	users     users.UserServiceClient
	gotenberg *gotenberg.Client
}

func NewServer(qc quote.QuoteServiceClient, uc users.UserServiceClient, gt *gotenberg.Client) *Server {
	return &Server{quote: qc, users: uc, gotenberg: gt}
}
