package tests

import (
	"context"
	authGrpc "gateway/auth"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type mockAuthServer struct {
	authGrpc.UnimplementedAuthServiceServer
}

func (s *mockAuthServer) Register(_ context.Context, _ *authGrpc.RegisterRequest) (*authGrpc.FormGenericResponse, error) {
	return &authGrpc.FormGenericResponse{Success: true}, nil
}

func (s *mockAuthServer) Login(_ context.Context, _ *authGrpc.LoginRequest) (*authGrpc.LoginResponse, error) {
	token := "test-token"
	return &authGrpc.LoginResponse{Success: true, Token: &token}, nil
}

func startTestServer(t *testing.T) (authGrpc.AuthServiceClient, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	authGrpc.RegisterAuthServiceServer(s, &mockAuthServer{})
	go s.Serve(lis)

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		s.Stop()
		t.Fatalf("failed to connect: %v", err)
	}

	return authGrpc.NewAuthServiceClient(conn), func() {
		conn.Close()
		s.Stop()
	}
}

func TestRegisterSuccess(t *testing.T) {
	client, cleanup := startTestServer(t)
	defer cleanup()

	resp, err := client.Register(context.Background(), &authGrpc.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}
	if !resp.GetSuccess() {
		t.Fatal("Expected success, got false")
	}
}

func TestLoginSuccess(t *testing.T) {
	client, cleanup := startTestServer(t)
	defer cleanup()

	resp, err := client.Login(context.Background(), &authGrpc.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	if !resp.GetSuccess() {
		t.Fatal("Expected success, got false")
	}
	if resp.GetToken() == "" {
		t.Fatal("Expected token, got empty string")
	}
}
