package actions

import (
	"database/sql"

	usersGrpc "project-devis-users/services/grpc"
)

type Server struct {
	usersGrpc.UnimplementedUserServiceServer
	db *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{db: db}
}
