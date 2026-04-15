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
	userClient userGrpc.AuthUserServiceClient
}

func NewServer(db *sql.DB, userConn *grpc.ClientConn) *Server {
	return &Server{
		db:         db,
		userClient: userGrpc.NewAuthUserServiceClient(userConn),
	}
}
