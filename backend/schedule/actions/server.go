package actions

import (
	"database/sql"

	scheduleGrpc "project-devis-schedule/services/grpc"
)

type Server struct {
	scheduleGrpc.UnimplementedScheduleServiceServer
	db *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{db: db}
}
