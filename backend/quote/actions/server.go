package actions

import (
	"database/sql"

	quoteGrpc "project-devis-quote/services/grpc"
)

type Server struct {
	quoteGrpc.UnimplementedQuoteServiceServer
	db *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{db: db}
}
