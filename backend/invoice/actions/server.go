package actions

import (
	"database/sql"

	"project-devis-invoice/pdp"
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
	pdpClient      pdp.Client
	pdpDirectory   pdp.Directory
}

func NewServer(
	db *sql.DB,
	quoteClient quoteGrpc.QuoteServiceClient,
	usersClient usersGrpc.UserServiceClient,
	scheduleClient scheduleGrpc.ScheduleServiceClient,
	pdpClient pdp.Client,
	pdpDirectory pdp.Directory,
) *Server {
	if pdpClient == nil { // tolerate nil (e.g. tests that don't exercise deposit)
		pdpClient = pdp.NoopClient{}
	}
	if pdpDirectory == nil {
		pdpDirectory = pdp.NoopDirectory{}
	}
	return &Server{
		db:             db,
		quoteClient:    quoteClient,
		usersClient:    usersClient,
		scheduleClient: scheduleClient,
		pdpClient:      pdpClient,
		pdpDirectory:   pdpDirectory,
	}
}
