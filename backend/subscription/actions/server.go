package actions

import (
	"database/sql"

	subscriptionGrpc "project-devis-subscription/services/grpc"
)

type Server struct {
	subscriptionGrpc.UnimplementedSubscriptionServiceServer
	db            *sql.DB
	webhookSecret string
}

func NewServer(db *sql.DB, webhookSecret string) *Server {
	return &Server{db: db, webhookSecret: webhookSecret}
}
