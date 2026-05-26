package tests

import (
	"context"
	"fmt"

	userGrpc "project-devis-auth/services/user_auth"

	"google.golang.org/grpc"
)

// MockUserClient implements userGrpc.UserServiceClient for testing.
type MockUserClient struct {
	CreateUserFn func(ctx context.Context, req *userGrpc.CreateUserRequest) (*userGrpc.CreateUserResponse, error)
	DeleteUserFn func(ctx context.Context, req *userGrpc.DeleteUserRequest) (*userGrpc.GenericResponse, error)
}

func (m *MockUserClient) CreateUser(ctx context.Context, in *userGrpc.CreateUserRequest, opts ...grpc.CallOption) (*userGrpc.CreateUserResponse, error) {
	if m.CreateUserFn != nil {
		return m.CreateUserFn(ctx, in)
	}
	return nil, fmt.Errorf("CreateUserFn not set")
}

func (m *MockUserClient) DeleteUser(ctx context.Context, in *userGrpc.DeleteUserRequest, opts ...grpc.CallOption) (*userGrpc.GenericResponse, error) {
	if m.DeleteUserFn != nil {
		return m.DeleteUserFn(ctx, in)
	}
	return nil, fmt.Errorf("DeleteUserFn not set")
}
