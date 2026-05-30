package grpc

import (
	"context"

	grpcpkg "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ScheduleServiceServer interface {
	mustEmbedUnimplementedScheduleServiceServer()
}

type UnimplementedScheduleServiceServer struct{}

func (UnimplementedScheduleServiceServer) mustEmbedUnimplementedScheduleServiceServer() {}

func RegisterScheduleServiceServer(s grpcpkg.ServiceRegistrar, srv ScheduleServiceServer) {
	s.RegisterService(&grpcpkg.ServiceDesc{
		ServiceName: "schedule.ScheduleService",
		HandlerType: (*ScheduleServiceServer)(nil),
		Methods:     []grpcpkg.MethodDesc{},
		Streams:     []grpcpkg.StreamDesc{},
		Metadata:    "schedule.proto",
	}, srv)
}

func UnimplementedError() error {
	return status.Error(codes.Unimplemented, "method not implemented")
}

type GenericResponse struct {
	Success bool
	Code    int32
}

type CreateScheduleRequest struct {
	UserId         string
	QuoteId        string
	Name           string
	StartMonth     string
	DurationMonths int32
}

type CreateScheduleResponse struct {
	Success    bool
	Code       int32
	ScheduleId string
}

type UpdateScheduleCellRequest struct {
	ScheduleId  string
	UserId      string
	QuoteLineId string
	MonthIndex  int32
	AmountEur   string
}

type ValidateScheduleRequest struct {
	ScheduleId string
	UserId     string
}

var _ context.Context