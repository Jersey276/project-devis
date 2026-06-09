package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net"

	"project-devis-email/actions"
	"project-devis-email/services"
	emailGrpc "project-devis-email/services/grpc"

	"google.golang.org/grpc"
)

//go:embed migrations
var migrationsFS embed.FS

var port = flag.Int("port", 50058, "The server port")

func main() {
	flag.Parse()

	db := services.ConnectDB()
	services.RunMigrations(db, migrationsFS)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.MaxRecvMsgSize(20 * 1024 * 1024))
	emailGrpc.RegisterEmailServiceServer(grpcServer, actions.NewServer(db))

	log.Printf("email gRPC server listening on %s", lis.Addr().String())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve failed: %v", err)
	}
}
