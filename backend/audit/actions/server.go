package actions

import (
	"database/sql"

	auditGrpc "project-devis-audit/services/grpc"
)

type Server struct {
	auditGrpc.UnimplementedAuditServiceServer
	db      *sql.DB
	purgeDB *sql.DB // DELETE-only connection for 6-month expiry; may be nil
}

func NewServer(db, purgeDB *sql.DB) *Server {
	return &Server{db: db, purgeDB: purgeDB}
}
