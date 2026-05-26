package actions

import (
	"database/sql"

	templateGrpc "project-devis-template/services/grpc"
)

type Server struct {
	templateGrpc.UnimplementedTemplateServiceServer
	db *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{db: db}
}
