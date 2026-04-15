package main

import (
	"context"
	"flag"
	"net/http"
	authGrpc "project-devis-auth/services/grpc"
	"testing"

	"google.golang.org/grpc"
)

	var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	authGrpc.UnimplementedAuthServiceServer
}


func TestRegister(t *testing.T) {
	// Create a test server using the main function
	ts := http.Server{
		Addr: ":50051",
	}

	if err := ts.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		t.Fatalf("Could not start test server: %v", err)
	}

	conn, err := grpc.NewClient("localhost:50051", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()
	client := authGrpc.NewAuthServiceClient(conn)

	resp, err := client.Register(context.Background(), &authGrpc.RegisterRequest{
		Email: "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}
	if !resp.Success {
		t.Fatalf("Registration failed: %v", resp.Message)
	}
}

func TestLogin(t *testing.T) {
	// Similar setup as TestRegister, but call the Login method instead
	conn, err := grpc.NewClient("localhost:50051", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()
	client := authGrpc.NewAuthServiceClient(conn)

	resp, err := client.Login(context.Background(), &authGrpc.LoginRequest{
		Email: "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	if !resp.Success {
		t.Fatalf("Login failed: %v", resp.Message)
	}
}
