package actions

import (
	"database/sql"

	projectGrpc "project-devis-project/services/grpc"
)

type Server struct {
	projectGrpc.UnimplementedProjectServiceServer
	db *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{db: db}
}
