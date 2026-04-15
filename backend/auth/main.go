package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"project-devis-auth/actions"
	"project-devis-auth/services"
	authGrpc "project-devis-auth/services/grpc"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

func main() {
	flag.Parse()

	db := services.ConnectDB()
	services.RunMigrations(db)
	userServiceAddress := services.UserServiceAddress.GetValue()
	if userServiceAddress == "" {
		userServiceAddress = "localhost:50052"
	}
	userConn, err := grpc.NewClient(userServiceAddress)
	if err != nil {
		log.Fatalf("failed to create user service client: %v", err)
	}
	defer userConn.Close()

	authServer := actions.NewServer(db, userConn)

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	authGrpc.RegisterAuthServiceServer(grpcServer, authServer)

	log.Printf("auth gRPC server listening on %s", lis.Addr().String())
	grpcServer.Serve(lis)
}
