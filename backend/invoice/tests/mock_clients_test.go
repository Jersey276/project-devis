package tests

import (
	"context"

	"google.golang.org/grpc"

	quoteGrpc "project-devis-invoice/services/quotegrpc"
	scheduleGrpc "project-devis-invoice/services/schedulegrpc"
	usersGrpc "project-devis-invoice/services/usersgrpc"
)

// mockQuoteClient implements quoteGrpc.QuoteServiceClient. Embedding the
// interface satisfies methods invoice never calls; only GetQuote is
// overridden since that's the only one actions/resolve.go and
// actions/create_from_quote.go use.
type mockQuoteClient struct {
	quoteGrpc.QuoteServiceClient
	GetQuoteFn func(ctx context.Context, in *quoteGrpc.GetQuoteRequest, opts ...grpc.CallOption) (*quoteGrpc.GetQuoteResponse, error)
}

func (m *mockQuoteClient) GetQuote(ctx context.Context, in *quoteGrpc.GetQuoteRequest, opts ...grpc.CallOption) (*quoteGrpc.GetQuoteResponse, error) {
	return m.GetQuoteFn(ctx, in, opts...)
}

// mockScheduleClient implements scheduleGrpc.ScheduleServiceClient, overriding
// only the methods invoice actually calls (GetSchedule, GetScheduleCells,
// ListSchedules).
type mockScheduleClient struct {
	scheduleGrpc.ScheduleServiceClient
	GetScheduleFn      func(ctx context.Context, in *scheduleGrpc.GetScheduleRequest, opts ...grpc.CallOption) (*scheduleGrpc.GetScheduleResponse, error)
	GetScheduleCellsFn func(ctx context.Context, in *scheduleGrpc.GetScheduleCellsRequest, opts ...grpc.CallOption) (*scheduleGrpc.GetScheduleCellsResponse, error)
	ListSchedulesFn    func(ctx context.Context, in *scheduleGrpc.ListSchedulesRequest, opts ...grpc.CallOption) (*scheduleGrpc.ListSchedulesResponse, error)
}

func (m *mockScheduleClient) GetSchedule(ctx context.Context, in *scheduleGrpc.GetScheduleRequest, opts ...grpc.CallOption) (*scheduleGrpc.GetScheduleResponse, error) {
	return m.GetScheduleFn(ctx, in, opts...)
}

func (m *mockScheduleClient) GetScheduleCells(ctx context.Context, in *scheduleGrpc.GetScheduleCellsRequest, opts ...grpc.CallOption) (*scheduleGrpc.GetScheduleCellsResponse, error) {
	return m.GetScheduleCellsFn(ctx, in, opts...)
}

func (m *mockScheduleClient) ListSchedules(ctx context.Context, in *scheduleGrpc.ListSchedulesRequest, opts ...grpc.CallOption) (*scheduleGrpc.ListSchedulesResponse, error) {
	if m.ListSchedulesFn != nil {
		return m.ListSchedulesFn(ctx, in, opts...)
	}
	return &scheduleGrpc.ListSchedulesResponse{Success: true}, nil
}

// mockUsersClient implements usersGrpc.UserServiceClient, overriding only the
// methods invoice actually calls (GetUser, GetClient, GetAddress, GetCountry,
// GetTax, ListTaxesForCountry).
type mockUsersClient struct {
	usersGrpc.UserServiceClient
	GetUserFn             func(ctx context.Context, in *usersGrpc.GetUserRequest, opts ...grpc.CallOption) (*usersGrpc.GetUserResponse, error)
	GetClientFn           func(ctx context.Context, in *usersGrpc.GetClientRequest, opts ...grpc.CallOption) (*usersGrpc.GetClientResponse, error)
	GetAddressFn          func(ctx context.Context, in *usersGrpc.GetAddressRequest, opts ...grpc.CallOption) (*usersGrpc.GetAddressResponse, error)
	GetCountryFn          func(ctx context.Context, in *usersGrpc.GetCountryRequest, opts ...grpc.CallOption) (*usersGrpc.GetCountryResponse, error)
	GetTaxFn              func(ctx context.Context, in *usersGrpc.GetTaxRequest, opts ...grpc.CallOption) (*usersGrpc.GetTaxResponse, error)
	ListTaxesForCountryFn func(ctx context.Context, in *usersGrpc.ListTaxesForCountryRequest, opts ...grpc.CallOption) (*usersGrpc.ListTaxesResponse, error)
}

func (m *mockUsersClient) GetUser(ctx context.Context, in *usersGrpc.GetUserRequest, opts ...grpc.CallOption) (*usersGrpc.GetUserResponse, error) {
	return m.GetUserFn(ctx, in, opts...)
}

func (m *mockUsersClient) GetClient(ctx context.Context, in *usersGrpc.GetClientRequest, opts ...grpc.CallOption) (*usersGrpc.GetClientResponse, error) {
	return m.GetClientFn(ctx, in, opts...)
}

func (m *mockUsersClient) GetAddress(ctx context.Context, in *usersGrpc.GetAddressRequest, opts ...grpc.CallOption) (*usersGrpc.GetAddressResponse, error) {
	return m.GetAddressFn(ctx, in, opts...)
}

func (m *mockUsersClient) GetCountry(ctx context.Context, in *usersGrpc.GetCountryRequest, opts ...grpc.CallOption) (*usersGrpc.GetCountryResponse, error) {
	return m.GetCountryFn(ctx, in, opts...)
}

func (m *mockUsersClient) GetTax(ctx context.Context, in *usersGrpc.GetTaxRequest, opts ...grpc.CallOption) (*usersGrpc.GetTaxResponse, error) {
	if m.GetTaxFn != nil {
		return m.GetTaxFn(ctx, in, opts...)
	}
	return &usersGrpc.GetTaxResponse{Success: true, Tax: &usersGrpc.Tax{Id: in.TaxId, Rate: "20", Name: "TVA 20%"}}, nil
}

func (m *mockUsersClient) ListTaxesForCountry(ctx context.Context, in *usersGrpc.ListTaxesForCountryRequest, opts ...grpc.CallOption) (*usersGrpc.ListTaxesResponse, error) {
	return m.ListTaxesForCountryFn(ctx, in, opts...)
}
