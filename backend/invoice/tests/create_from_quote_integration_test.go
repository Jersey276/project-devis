package tests

import (
	"context"
	"testing"

	"google.golang.org/grpc"

	"project-devis-invoice/actions"
	invoiceGrpc "project-devis-invoice/services/grpc"
	quoteGrpc "project-devis-invoice/services/quotegrpc"
	scheduleGrpc "project-devis-invoice/services/schedulegrpc"
	usersGrpc "project-devis-invoice/services/usersgrpc"
)

func domesticFixtures() (*mockQuoteClient, *mockScheduleClient, *mockUsersClient) {
	quote := &mockQuoteClient{
		GetQuoteFn: func(_ context.Context, req *quoteGrpc.GetQuoteRequest, _ ...grpc.CallOption) (*quoteGrpc.GetQuoteResponse, error) {
			return &quoteGrpc.GetQuoteResponse{
				Success: true,
				Quote: &quoteGrpc.Quote{
					QuoteId:       req.QuoteId,
					UserId:        req.UserId,
					State:         quoteGrpc.QuoteState_QUOTE_STATE_VALIDATED,
					ClientId:      "client-1",
					AddressId:     10,
					UserAddressId: 20,
				},
				Lines: []*quoteGrpc.QuoteLine{
					{LineId: "line-1", QuoteId: req.QuoteId, Name: "Prestation", Unit: "u", Quantity: "1", UnitPrice: 10000, Position: 0, TaxId: 1},
				},
			}, nil
		},
	}
	schedule := &mockScheduleClient{
		ListSchedulesFn: func(_ context.Context, _ *scheduleGrpc.ListSchedulesRequest, _ ...grpc.CallOption) (*scheduleGrpc.ListSchedulesResponse, error) {
			return &scheduleGrpc.ListSchedulesResponse{Success: true}, nil
		},
	}
	users := &mockUsersClient{
		GetUserFn: func(_ context.Context, _ *usersGrpc.GetUserRequest, _ ...grpc.CallOption) (*usersGrpc.GetUserResponse, error) {
			return &usersGrpc.GetUserResponse{Success: true, User: &usersGrpc.User{
				UserId: "user-1", Company: "Acme SARL", Siren: "123456782", Vat: "FR12345678901", Email: "acme@example.com",
			}}, nil
		},
		GetClientFn: func(_ context.Context, _ *usersGrpc.GetClientRequest, _ ...grpc.CallOption) (*usersGrpc.GetClientResponse, error) {
			return &usersGrpc.GetClientResponse{Success: true, Client: &usersGrpc.Client{
				ClientId: "client-1", FirstName: "Jean", LastName: "Dupont", ClientType: usersGrpc.ClientType_CLIENT_TYPE_INDIVIDUAL,
			}}, nil
		},
		GetAddressFn: func(_ context.Context, req *usersGrpc.GetAddressRequest, _ ...grpc.CallOption) (*usersGrpc.GetAddressResponse, error) {
			if req.OwnerType == usersGrpc.OwnerType_OWNER_TYPE_USER {
				return &usersGrpc.GetAddressResponse{Success: true, Address: &usersGrpc.Address{
					Street: "1 rue de Paris", ZipCode: "75001", City: "Paris", CountryId: 1,
				}}, nil
			}
			return &usersGrpc.GetAddressResponse{Success: true, Address: &usersGrpc.Address{
				Street: "2 rue de Lyon", ZipCode: "69001", City: "Lyon", CountryId: 1,
			}}, nil
		},
		GetCountryFn: func(_ context.Context, req *usersGrpc.GetCountryRequest, _ ...grpc.CallOption) (*usersGrpc.GetCountryResponse, error) {
			return &usersGrpc.GetCountryResponse{Success: true, Country: &usersGrpc.Country{Id: req.CountryId, Code: "FR", IsEu: true}}, nil
		},
	}
	return quote, schedule, users
}

func TestCreateInvoiceFromQuote_IssueNow_Success(t *testing.T) {
	db := sealTestDB(t)
	quote, schedule, users := domesticFixtures()
	srv := actions.NewServer(db, quote, users, schedule, nil, nil, nil)

	resp, err := srv.CreateInvoiceFromQuote(context.Background(), &invoiceGrpc.CreateInvoiceFromQuoteRequest{
		UserId: "user-1", QuoteId: "quote-1", IssueNow: true,
	})
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}
	if !resp.Success || resp.InvoiceId == "" || resp.InvoiceNumber == "" {
		t.Fatalf("create invoice failed: success=%v code=%d number=%q", resp.Success, resp.Code, resp.InvoiceNumber)
	}

	get, err := srv.GetInvoice(context.Background(), &invoiceGrpc.GetInvoiceRequest{InvoiceId: resp.InvoiceId, UserId: "user-1"})
	if err != nil {
		t.Fatalf("get invoice: %v", err)
	}
	if !get.Success || get.Invoice.GetStatus() != "ISSUED" {
		t.Fatalf("expected ISSUED invoice, got status=%q code=%d", get.Invoice.GetStatus(), get.Code)
	}
	if got := get.Invoice.GetTotalHtCents(); got != 10000 {
		t.Errorf("total HT = %d; want 10000", got)
	}
	if got := get.Invoice.GetTotalTtcCents(); got != 12000 {
		t.Errorf("total TTC = %d; want 12000 (20%% VAT)", got)
	}
	if get.Invoice.GetClient().GetSiren() != "" {
		t.Errorf("client is an individual; SIREN should be empty, got %q", get.Invoice.GetClient().GetSiren())
	}
}

func TestCreateInvoiceFromQuote_Draft_ThenIssueSeparately(t *testing.T) {
	db := sealTestDB(t)
	quote, schedule, users := domesticFixtures()
	srv := actions.NewServer(db, quote, users, schedule, nil, nil, nil)

	created, err := srv.CreateInvoiceFromQuote(context.Background(), &invoiceGrpc.CreateInvoiceFromQuoteRequest{
		UserId: "user-1", QuoteId: "quote-1",
	})
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}
	if !created.Success || created.InvoiceNumber != "" {
		t.Fatalf("expected draft with no number, got success=%v number=%q", created.Success, created.InvoiceNumber)
	}

	preview, err := srv.GetInvoice(context.Background(), &invoiceGrpc.GetInvoiceRequest{InvoiceId: created.InvoiceId, UserId: "user-1"})
	if err != nil {
		t.Fatalf("get draft preview: %v", err)
	}
	if !preview.Success || preview.Invoice.GetStatus() != "DRAFT" {
		t.Fatalf("expected DRAFT preview, got status=%q", preview.Invoice.GetStatus())
	}
	if got := preview.Invoice.GetTotalTtcCents(); got != 12000 {
		t.Errorf("draft preview TTC = %d; want 12000", got)
	}

	issued, err := srv.IssueInvoice(context.Background(), &invoiceGrpc.IssueInvoiceRequest{InvoiceId: created.InvoiceId, UserId: "user-1"})
	if err != nil {
		t.Fatalf("issue invoice: %v", err)
	}
	if !issued.Success || issued.InvoiceNumber == "" {
		t.Fatalf("issue failed: success=%v code=%d", issued.Success, issued.Code)
	}

	// Re-issuing an already-issued invoice is idempotent, not an error.
	reissued, err := srv.IssueInvoice(context.Background(), &invoiceGrpc.IssueInvoiceRequest{InvoiceId: created.InvoiceId, UserId: "user-1"})
	if err != nil {
		t.Fatalf("re-issue invoice: %v", err)
	}
	if !reissued.Success || reissued.InvoiceNumber != issued.InvoiceNumber {
		t.Fatalf("re-issue should return the same number: first=%q second=%q", issued.InvoiceNumber, reissued.InvoiceNumber)
	}
}

func TestCreateInvoiceFromQuote_OSSApplies_UsesDestinationRate(t *testing.T) {
	db := sealTestDB(t)
	quote, schedule, users := domesticFixtures()

	// German B2C client, issuer opted into OSS: destination-country VAT (19%)
	// replaces the origin rate, per art. 259 D CGI.
	users.GetUserFn = func(_ context.Context, _ *usersGrpc.GetUserRequest, _ ...grpc.CallOption) (*usersGrpc.GetUserResponse, error) {
		return &usersGrpc.GetUserResponse{Success: true, User: &usersGrpc.User{
			UserId: "user-1", Company: "Acme SARL", Vat: "FR12345678901", OssEnabled: true,
		}}, nil
	}
	users.GetAddressFn = func(_ context.Context, req *usersGrpc.GetAddressRequest, _ ...grpc.CallOption) (*usersGrpc.GetAddressResponse, error) {
		if req.OwnerType == usersGrpc.OwnerType_OWNER_TYPE_USER {
			return &usersGrpc.GetAddressResponse{Success: true, Address: &usersGrpc.Address{Street: "1 rue de Paris", CountryId: 1}}, nil
		}
		return &usersGrpc.GetAddressResponse{Success: true, Address: &usersGrpc.Address{Street: "1 Hauptstrasse", CountryId: 2}}, nil
	}
	users.GetCountryFn = func(_ context.Context, req *usersGrpc.GetCountryRequest, _ ...grpc.CallOption) (*usersGrpc.GetCountryResponse, error) {
		if req.CountryId == 2 {
			return &usersGrpc.GetCountryResponse{Success: true, Country: &usersGrpc.Country{Id: 2, Code: "DE", IsEu: true}}, nil
		}
		return &usersGrpc.GetCountryResponse{Success: true, Country: &usersGrpc.Country{Id: 1, Code: "FR", IsEu: true}}, nil
	}
	users.ListTaxesForCountryFn = func(_ context.Context, _ *usersGrpc.ListTaxesForCountryRequest, _ ...grpc.CallOption) (*usersGrpc.ListTaxesResponse, error) {
		return &usersGrpc.ListTaxesResponse{Success: true, Taxes: []*usersGrpc.Tax{
			{Id: 99, Rate: "19", Name: "MwSt 19%", IsDefault: true},
		}}, nil
	}

	srv := actions.NewServer(db, quote, users, schedule, nil, nil, nil)
	resp, err := srv.CreateInvoiceFromQuote(context.Background(), &invoiceGrpc.CreateInvoiceFromQuoteRequest{
		UserId: "user-1", QuoteId: "quote-1", IssueNow: true,
	})
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}
	if !resp.Success {
		t.Fatalf("create invoice failed: code=%d", resp.Code)
	}

	get, err := srv.GetInvoice(context.Background(), &invoiceGrpc.GetInvoiceRequest{InvoiceId: resp.InvoiceId, UserId: "user-1"})
	if err != nil {
		t.Fatalf("get invoice: %v", err)
	}
	if !get.Invoice.GetOssApplied() {
		t.Fatalf("expected oss_applied=true")
	}
	if got := get.Invoice.GetTotalTtcCents(); got != 11900 {
		t.Errorf("total TTC = %d; want 11900 (19%% German VAT on 10000 HT)", got)
	}
}

func TestCreateInvoiceFromQuote_QuoteNotValidated_ReturnsSourceNotEligible(t *testing.T) {
	db := sealTestDB(t)
	quote, schedule, users := domesticFixtures()
	quote.GetQuoteFn = func(_ context.Context, req *quoteGrpc.GetQuoteRequest, _ ...grpc.CallOption) (*quoteGrpc.GetQuoteResponse, error) {
		return &quoteGrpc.GetQuoteResponse{Success: true, Quote: &quoteGrpc.Quote{
			QuoteId: req.QuoteId, UserId: req.UserId, State: quoteGrpc.QuoteState_QUOTE_STATE_DRAFT,
		}}, nil
	}
	srv := actions.NewServer(db, quote, users, schedule, nil, nil, nil)

	resp, err := srv.CreateInvoiceFromQuote(context.Background(), &invoiceGrpc.CreateInvoiceFromQuoteRequest{
		UserId: "user-1", QuoteId: "quote-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for a non-validated quote")
	}
}

func TestCreateInvoiceFromQuote_QuoteAlreadyHasSchedule_Refused(t *testing.T) {
	db := sealTestDB(t)
	quote, schedule, users := domesticFixtures()
	schedule.ListSchedulesFn = func(_ context.Context, _ *scheduleGrpc.ListSchedulesRequest, _ ...grpc.CallOption) (*scheduleGrpc.ListSchedulesResponse, error) {
		return &scheduleGrpc.ListSchedulesResponse{Success: true, Schedules: []*scheduleGrpc.ScheduleSummary{{ScheduleId: "sched-1"}}}, nil
	}
	srv := actions.NewServer(db, quote, users, schedule, nil, nil, nil)

	resp, err := srv.CreateInvoiceFromQuote(context.Background(), &invoiceGrpc.CreateInvoiceFromQuoteRequest{
		UserId: "user-1", QuoteId: "quote-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure when the quote already has a schedule")
	}
}

func TestCreateInvoiceFromQuote_MissingFields_InvalidInput(t *testing.T) {
	db := sealTestDB(t)
	quote, schedule, users := domesticFixtures()
	srv := actions.NewServer(db, quote, users, schedule, nil, nil, nil)

	resp, err := srv.CreateInvoiceFromQuote(context.Background(), &invoiceGrpc.CreateInvoiceFromQuoteRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success || len(resp.ValidationErrors) != 2 {
		t.Fatalf("expected 2 validation errors (user_id, quote_id), got %d", len(resp.ValidationErrors))
	}
}
