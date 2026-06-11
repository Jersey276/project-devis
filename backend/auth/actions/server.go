package actions

import (
	"database/sql"
	"project-devis-auth/services"
	authGrpc "project-devis-auth/services/grpc"
	userGrpc "project-devis-auth/services/user_auth"

	"google.golang.org/grpc"
)

type Server struct {
	authGrpc.UnimplementedAuthServiceServer
	db          *sql.DB
	userClient  userGrpc.UserServiceClient
	emailSender services.EmailSender
}

func NewServer(db *sql.DB, userConn *grpc.ClientConn) *Server {
	return &Server{
		db:          db,
		userClient:  userGrpc.NewUserServiceClient(userConn),
		emailSender: services.NewEmailSenderFromEnv(),
	}
}

func NewServerWithClient(db *sql.DB, userClient userGrpc.UserServiceClient) *Server {
	return &Server{
		db:          db,
		userClient:  userClient,
		emailSender: services.NewEmailSenderFromEnv(),
	}
}
