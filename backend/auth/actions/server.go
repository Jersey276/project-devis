package actions

import (
	"database/sql"
	authGrpc "project-devis-auth/services/grpc"
	userGrpc "project-devis-auth/services/user_auth"

	"google.golang.org/grpc"
)

type Server struct {
	authGrpc.UnimplementedAuthServiceServer
	db         *sql.DB
	userClient userGrpc.UserServiceClient
}

func NewServer(db *sql.DB, userConn *grpc.ClientConn) *Server {
	return &Server{
		db:         db,
		userClient: userGrpc.NewUserServiceClient(userConn),
	}
}

func NewServerWithClient(db *sql.DB, userClient userGrpc.UserServiceClient) *Server {
	return &Server{
		db:         db,
		userClient: userClient,
	}
}
