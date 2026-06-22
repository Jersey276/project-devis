package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net"

	"project-devis-audit/actions"
	"project-devis-audit/services"
	auditGrpc "project-devis-audit/services/grpc"

	"google.golang.org/grpc"
)

//go:embed migrations
var migrationsFS embed.FS

var port = flag.Int("port", 50060, "The server port")

func main() {
	flag.Parse()

	db := services.ConnectDB()
	services.RunMigrations(db, migrationsFS)

	purgeDB := services.ConnectPurgeDB()
	actions.StartPurgeWorker(purgeDB)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	auditGrpc.RegisterAuditServiceServer(grpcServer, actions.NewServer(db, purgeDB))

	log.Printf("audit gRPC server listening on %s", lis.Addr().String())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve failed: %v", err)
	}
}
