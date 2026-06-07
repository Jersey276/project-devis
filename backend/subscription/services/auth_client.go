package services

import (
	"os"
	"sync"

	authpb "project-devis-subscription/services/grpc/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	authClientOnce sync.Once
	authClientInst authpb.AuthServiceClient
	authClientErr  error
)

func GetAuthServiceClient() (authpb.AuthServiceClient, error) {
	authClientOnce.Do(func() {
		address := os.Getenv("AUTH_SERVICE_ADDRESS")
		if address == "" {
			address = "localhost:50051"
		}
		conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			authClientErr = err
			return
		}
		authClientInst = authpb.NewAuthServiceClient(conn)
	})
	return authClientInst, authClientErr
}
