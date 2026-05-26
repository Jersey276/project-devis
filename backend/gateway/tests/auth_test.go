package tests

import (
	"context"
	"flag"
	authGrpc "gateway/auth"
	"net/http"
	"testing"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	authGrpc.UnimplementedAuthServiceServer
}

func TestRegisterSuccess(t *testing.T) {
	// Create a test server using the main function
	ts := http.Server{
		Addr: ":50051",
	}

	ts.ListenAndServe()

	conn, err := grpc.NewClient("localhost:50051", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()
	client := authGrpc.NewAuthServiceClient(conn)
	resp, err := client.Register(context.Background(), &authGrpc.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}
	if resp.GetSuccess() != true {
		t.Fatalf("Expected success, got %v", resp.GetSuccess())
	}
}

func TestLoginSuccess(t *testing.T) {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()
	client := authGrpc.NewAuthServiceClient(conn)
	resp, err := client.Login(context.Background(), &authGrpc.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	if resp.GetSuccess() != true {
		t.Fatalf("Expected success, got %v", resp.GetSuccess())
	}
	if resp.GetToken() == "" {
		t.Fatalf("Expected token, got empty string")
	}
}
