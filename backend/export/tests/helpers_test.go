package tests

import (
	"context"

	"google.golang.org/grpc"

	"project-devis-export/quote"
	schedulepb "project-devis-export/services/schedule"
	"project-devis-export/users"
)

// fakeQuote embeds the generated client interface so unimplemented methods
// (e.g. CreateQuote) cause a nil-dereference panic if the orchestrator
// accidentally calls them — making the test surface explicit.
type fakeQuote struct {
	quote.QuoteServiceClient
	getQuote func(context.Context, *quote.GetQuoteRequest) (*quote.GetQuoteResponse, error)
}

func (f *fakeQuote) GetQuote(ctx context.Context, in *quote.GetQuoteRequest, _ ...grpc.CallOption) (*quote.GetQuoteResponse, error) {
	return f.getQuote(ctx, in)
}

type fakeUsers struct {
	users.UserServiceClient
	getUser          func(context.Context, *users.GetUserRequest) (*users.GetUserResponse, error)
	listAddresses    func(context.Context, *users.ListAddressesRequest) (*users.ListAddressesResponse, error)
	getClient        func(context.Context, *users.GetClientRequest) (*users.GetClientResponse, error)
	getAddress       func(context.Context, *users.GetAddressRequest) (*users.GetAddressResponse, error)
	listTaxesForUser func(context.Context, *users.ListTaxesForUserRequest) (*users.ListTaxesResponse, error)
}

type fakeSchedule struct {
	schedulepb.ScheduleServiceClient
	getSchedule func(context.Context, *schedulepb.GetScheduleRequest) (*schedulepb.GetScheduleResponse, error)
}

func (f *fakeSchedule) GetSchedule(ctx context.Context, in *schedulepb.GetScheduleRequest, _ ...grpc.CallOption) (*schedulepb.GetScheduleResponse, error) {
	return f.getSchedule(ctx, in)
}

func (f *fakeUsers) GetUser(ctx context.Context, in *users.GetUserRequest, _ ...grpc.CallOption) (*users.GetUserResponse, error) {
	return f.getUser(ctx, in)
}

func (f *fakeUsers) ListAddresses(ctx context.Context, in *users.ListAddressesRequest, _ ...grpc.CallOption) (*users.ListAddressesResponse, error) {
	return f.listAddresses(ctx, in)
}

func (f *fakeUsers) GetClient(ctx context.Context, in *users.GetClientRequest, _ ...grpc.CallOption) (*users.GetClientResponse, error) {
	return f.getClient(ctx, in)
}

func (f *fakeUsers) GetAddress(ctx context.Context, in *users.GetAddressRequest, _ ...grpc.CallOption) (*users.GetAddressResponse, error) {
	return f.getAddress(ctx, in)
}

func (f *fakeUsers) ListTaxesForUser(ctx context.Context, in *users.ListTaxesForUserRequest, _ ...grpc.CallOption) (*users.ListTaxesResponse, error) {
	if f.listTaxesForUser != nil {
		return f.listTaxesForUser(ctx, in)
	}
	return &users.ListTaxesResponse{Success: true}, nil
}

// fakeGotenberg satisfies the orchestrator's unexported pdfConverter interface
// (structural typing — Convert + ConvertPDFA3). Call counters let tests assert
// which path ran (e.g. that a draft never reaches PDF/A rendering).
type fakeGotenberg struct {
	convert      func(context.Context, []byte) ([]byte, error)
	convertPDFA3 func(context.Context, []byte) ([]byte, error)
	convertCalls int
	pdfa3Calls   int
}

func (f *fakeGotenberg) Convert(ctx context.Context, html []byte) ([]byte, error) {
	f.convertCalls++
	return f.convert(ctx, html)
}

func (f *fakeGotenberg) ConvertPDFA3(ctx context.Context, html []byte) ([]byte, error) {
	f.pdfa3Calls++
	if f.convertPDFA3 != nil {
		return f.convertPDFA3(ctx, html)
	}
	return f.convert(ctx, html)
}

// happyFakes builds a set of fakes that exercise the full Export pipeline
// without errors. Individual tests override fields to inject failures.
func happyFakes() (*fakeQuote, *fakeUsers, *fakeGotenberg) {
	qc := &fakeQuote{
		getQuote: func(_ context.Context, req *quote.GetQuoteRequest) (*quote.GetQuoteResponse, error) {
			return &quote.GetQuoteResponse{
				Success: true,
				Quote: &quote.Quote{
					QuoteId:   req.QuoteId,
					UserId:    req.UserId,
					Name:      "Cuisine équipée",
					ClientId:  "client-1",
					AddressId: 42,
				},
				Lines: []*quote.QuoteLine{
					{LineId: "l1", Name: "Pose plan de travail", Quantity: "1", UnitPrice: 50000, TaxId: 1},
				},
			}, nil
		},
	}
	uc := &fakeUsers{
		getUser: func(_ context.Context, req *users.GetUserRequest) (*users.GetUserResponse, error) {
			return &users.GetUserResponse{
				Success: true,
				User:    &users.User{UserId: req.UserId, Email: "me@example.com", Company: "Ateliers Martin"},
			}, nil
		},
		listAddresses: func(context.Context, *users.ListAddressesRequest) (*users.ListAddressesResponse, error) {
			return &users.ListAddressesResponse{
				Success:   true,
				Addresses: []*users.Address{{Id: 1, Street: "1 rue Test", City: "Paris", ZipCode: "75001"}},
			}, nil
		},
		getClient: func(_ context.Context, req *users.GetClientRequest) (*users.GetClientResponse, error) {
			return &users.GetClientResponse{
				Success: true,
				Client:  &users.Client{ClientId: req.ClientId, FirstName: "Marie", LastName: "Durand"},
			}, nil
		},
		getAddress: func(_ context.Context, req *users.GetAddressRequest) (*users.GetAddressResponse, error) {
			return &users.GetAddressResponse{
				Success: true,
				Address: &users.Address{Id: req.AddressId, Street: "2 av Recipient", City: "Lyon", ZipCode: "69000"},
			}, nil
		},
		listTaxesForUser: func(_ context.Context, req *users.ListTaxesForUserRequest) (*users.ListTaxesResponse, error) {
			taxes := make([]*users.Tax, 0, len(req.IncludeIds))
			for _, id := range req.IncludeIds {
				taxes = append(taxes, &users.Tax{Id: id, Name: "TVA 20%", Rate: "20.00"})
			}
			return &users.ListTaxesResponse{Success: true, Taxes: taxes}, nil
		},
	}
	gt := &fakeGotenberg{
		convert: func(context.Context, []byte) ([]byte, error) {
			return []byte("%PDF-1.4 fake"), nil
		},
	}
	return qc, uc, gt
}
