package tests

import (
	"context"
	"fmt"

	userGrpc "project-devis-auth/services/user_auth"

	"google.golang.org/grpc"
)

// MockUserClient implements userGrpc.AuthUserServiceClient for testing.
type MockUserClient struct {
	InsertUserFn func(ctx context.Context, req *userGrpc.InsertUserRequest) (*userGrpc.InsertUserResponse, error)
	DeleteUserFn func(ctx context.Context, req *userGrpc.DeleteUserRequest) (*userGrpc.GenericResponse, error)
}

func (m *MockUserClient) InsertUser(ctx context.Context, in *userGrpc.InsertUserRequest, opts ...grpc.CallOption) (*userGrpc.InsertUserResponse, error) {
	if m.InsertUserFn != nil {
		return m.InsertUserFn(ctx, in)
	}
	return nil, fmt.Errorf("InsertUserFn not set")
}

func (m *MockUserClient) DeleteUser(ctx context.Context, in *userGrpc.DeleteUserRequest, opts ...grpc.CallOption) (*userGrpc.GenericResponse, error) {
	if m.DeleteUserFn != nil {
		return m.DeleteUserFn(ctx, in)
	}
	return nil, fmt.Errorf("DeleteUserFn not set")
}
