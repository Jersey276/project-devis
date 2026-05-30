package tests

import (
	"context"
	"fmt"

	userGrpc "project-devis-auth/services/user_auth"

	"google.golang.org/grpc"
)

// MockUserClient implements userGrpc.UserServiceClient for testing.
type MockUserClient struct {
	CreateUserFn               func(ctx context.Context, req *userGrpc.CreateUserRequest) (*userGrpc.CreateUserResponse, error)
	DeleteUserFn               func(ctx context.Context, req *userGrpc.DeleteUserRequest) (*userGrpc.GenericResponse, error)
	GetUserAccessInfoFn        func(ctx context.Context, req *userGrpc.GetUserAccessInfoRequest) (*userGrpc.GetUserAccessInfoResponse, error)
	GetUserAccessInfoByEmailFn func(ctx context.Context, req *userGrpc.GetUserAccessInfoByEmailRequest) (*userGrpc.GetUserAccessInfoResponse, error)
	TouchUserLastLoginFn       func(ctx context.Context, req *userGrpc.TouchUserLastLoginRequest) (*userGrpc.GenericResponse, error)
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

func (m *MockUserClient) GetUserAccessInfo(ctx context.Context, in *userGrpc.GetUserAccessInfoRequest, opts ...grpc.CallOption) (*userGrpc.GetUserAccessInfoResponse, error) {
	if m.GetUserAccessInfoFn != nil {
		return m.GetUserAccessInfoFn(ctx, in)
	}
	return &userGrpc.GetUserAccessInfoResponse{Success: true, Code: 0, UserId: in.GetUserId(), Email: "test@test.fr", Role: "user", Suspended: false}, nil
}

func (m *MockUserClient) GetUserAccessInfoByEmail(ctx context.Context, in *userGrpc.GetUserAccessInfoByEmailRequest, opts ...grpc.CallOption) (*userGrpc.GetUserAccessInfoResponse, error) {
	if m.GetUserAccessInfoByEmailFn != nil {
		return m.GetUserAccessInfoByEmailFn(ctx, in)
	}
	return &userGrpc.GetUserAccessInfoResponse{Success: true, Code: 0, UserId: "user-123", Email: in.GetEmail(), Role: "user", Suspended: false}, nil
}

func (m *MockUserClient) TouchUserLastLogin(ctx context.Context, in *userGrpc.TouchUserLastLoginRequest, opts ...grpc.CallOption) (*userGrpc.GenericResponse, error) {
	if m.TouchUserLastLoginFn != nil {
		return m.TouchUserLastLoginFn(ctx, in)
	}
	return &userGrpc.GenericResponse{Success: true, Code: 0}, nil
}
