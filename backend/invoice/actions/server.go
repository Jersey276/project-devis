package actions

import (
	"database/sql"

	invoiceGrpc "project-devis-invoice/services/grpc"
	quoteGrpc "project-devis-invoice/services/quotegrpc"
	scheduleGrpc "project-devis-invoice/services/schedulegrpc"
	usersGrpc "project-devis-invoice/services/usersgrpc"
)

type Server struct {
	invoiceGrpc.UnimplementedInvoiceServiceServer
	db             *sql.DB
	quoteClient    quoteGrpc.QuoteServiceClient
	usersClient    usersGrpc.UserServiceClient
	scheduleClient scheduleGrpc.ScheduleServiceClient
}

func NewServer(
	db *sql.DB,
	quoteClient quoteGrpc.QuoteServiceClient,
	usersClient usersGrpc.UserServiceClient,
	scheduleClient scheduleGrpc.ScheduleServiceClient,
) *Server {
	return &Server{
		db:             db,
		quoteClient:    quoteClient,
		usersClient:    usersClient,
		scheduleClient: scheduleClient,
	}
}
