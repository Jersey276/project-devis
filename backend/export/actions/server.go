package actions

import (
	"project-devis-export/quote"
	"project-devis-export/services/gotenberg"
	exportGrpc "project-devis-export/services/grpc"
	"project-devis-export/services/schedule"
	"project-devis-export/users"
)

type Server struct {
	exportGrpc.UnimplementedExportServiceServer
	quote     quote.QuoteServiceClient
	users     users.UserServiceClient
	schedule  schedule.ScheduleServiceClient
	gotenberg *gotenberg.Client
}

func NewServer(qc quote.QuoteServiceClient, uc users.UserServiceClient, sc schedule.ScheduleServiceClient, gt *gotenberg.Client) *Server {
	return &Server{quote: qc, users: uc, schedule: sc, gotenberg: gt}
}
