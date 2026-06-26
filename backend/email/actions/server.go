package actions

import (
	"database/sql"

	"project-devis-email/actions/codes"
	"project-devis-email/services"
	emailGrpc "project-devis-email/services/grpc"
)

const (
	CodeSuccess       = codes.Success
	CodeNotFound      = codes.NotFound
	CodeInternalError = codes.InternalError
	CodeInvalidInput  = codes.InvalidInput
)

type Server struct {
	emailGrpc.UnimplementedEmailServiceServer
	db     *sql.DB
	sender services.EmailSender
}

func NewServer(db *sql.DB) *Server {
	return &Server{
		db:     db,
		sender: services.NewEmailSenderFromEnv(),
	}
}
