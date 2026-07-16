package tests

import (
	"context"
	"testing"

	"google.golang.org/grpc"

	"project-devis-invoice/actions"
	invoiceGrpc "project-devis-invoice/services/grpc"
	scheduleGrpc "project-devis-invoice/services/schedulegrpc"
)

func scheduleFixtures() (*mockQuoteClient, *mockScheduleClient, *mockUsersClient) {
	quote, _, users := domesticFixtures()
	schedule := &mockScheduleClient{
		GetScheduleFn: func(_ context.Context, req *scheduleGrpc.GetScheduleRequest, _ ...grpc.CallOption) (*scheduleGrpc.GetScheduleResponse, error) {
			return &scheduleGrpc.GetScheduleResponse{Success: true, Schedule: &scheduleGrpc.ScheduleDetails{
				ScheduleId: req.ScheduleId, QuoteId: "quote-1", Status: "VALID", DurationMonths: 3,
			}}, nil
		},
		GetScheduleCellsFn: func(_ context.Context, _ *scheduleGrpc.GetScheduleCellsRequest, _ ...grpc.CallOption) (*scheduleGrpc.GetScheduleCellsResponse, error) {
			return &scheduleGrpc.GetScheduleCellsResponse{Success: true, Cells: []*scheduleGrpc.ScheduleCell{
				{QuoteLineId: "line-1", MonthIndex: 1, AmountCents: 3000},
				{QuoteLineId: "line-1", MonthIndex: 2, AmountCents: 3500},
				{QuoteLineId: "line-1", MonthIndex: 3, AmountCents: 3500},
			}}, nil
		},
	}
	return quote, schedule, users
}

func TestCreateInvoiceFromSchedule_IssueNow_BillsSelectedMonths(t *testing.T) {
	db := sealTestDB(t)
	quote, schedule, users := scheduleFixtures()
	srv := actions.NewServer(db, quote, users, schedule, nil, nil, nil)

	resp, err := srv.CreateInvoiceFromSchedule(context.Background(), &invoiceGrpc.CreateInvoiceFromScheduleRequest{
		UserId: "user-1", ScheduleId: "sched-1", MonthIndexes: []int32{1, 2}, IssueNow: true,
	})
	if err != nil {
		t.Fatalf("create invoice from schedule: %v", err)
	}
	if !resp.Success || resp.InvoiceNumber == "" {
		t.Fatalf("create failed: success=%v code=%d", resp.Success, resp.Code)
	}

	get, err := srv.GetInvoice(context.Background(), &invoiceGrpc.GetInvoiceRequest{InvoiceId: resp.InvoiceId, UserId: "user-1"})
	if err != nil {
		t.Fatalf("get invoice: %v", err)
	}
	// Months 1+2 = 3000+3500 = 6500 HT, at 20% domestic VAT -> 7800 TTC.
	if got := get.Invoice.GetTotalHtCents(); got != 6500 {
		t.Errorf("total HT = %d; want 6500 (months 1+2 only)", got)
	}
	if got := get.Invoice.GetTotalTtcCents(); got != 7800 {
		t.Errorf("total TTC = %d; want 7800", got)
	}
}

func TestCreateInvoiceFromSchedule_RebillingSameMonth_Refused(t *testing.T) {
	db := sealTestDB(t)
	quote, schedule, users := scheduleFixtures()
	srv := actions.NewServer(db, quote, users, schedule, nil, nil, nil)

	first, err := srv.CreateInvoiceFromSchedule(context.Background(), &invoiceGrpc.CreateInvoiceFromScheduleRequest{
		UserId: "user-1", ScheduleId: "sched-1", MonthIndexes: []int32{1}, IssueNow: true,
	})
	if err != nil || !first.Success {
		t.Fatalf("first invoice failed: err=%v success=%v code=%d", err, first.Success, first.Code)
	}

	second, err := srv.CreateInvoiceFromSchedule(context.Background(), &invoiceGrpc.CreateInvoiceFromScheduleRequest{
		UserId: "user-1", ScheduleId: "sched-1", MonthIndexes: []int32{1}, IssueNow: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if second.Success {
		t.Fatal("expected refusal when re-billing an already-billed month")
	}
}

func TestCreateInvoiceFromSchedule_NotValidStatus_Refused(t *testing.T) {
	db := sealTestDB(t)
	quote, schedule, users := scheduleFixtures()
	schedule.GetScheduleFn = func(_ context.Context, req *scheduleGrpc.GetScheduleRequest, _ ...grpc.CallOption) (*scheduleGrpc.GetScheduleResponse, error) {
		return &scheduleGrpc.GetScheduleResponse{Success: true, Schedule: &scheduleGrpc.ScheduleDetails{
			ScheduleId: req.ScheduleId, QuoteId: "quote-1", Status: "DRAFT", DurationMonths: 3,
		}}, nil
	}
	srv := actions.NewServer(db, quote, users, schedule, nil, nil, nil)

	resp, err := srv.CreateInvoiceFromSchedule(context.Background(), &invoiceGrpc.CreateInvoiceFromScheduleRequest{
		UserId: "user-1", ScheduleId: "sched-1", MonthIndexes: []int32{1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected refusal for a non-VALID schedule")
	}
}

func TestCreateInvoiceFromSchedule_NoMonthsSelected_InvalidInput(t *testing.T) {
	db := sealTestDB(t)
	quote, schedule, users := scheduleFixtures()
	srv := actions.NewServer(db, quote, users, schedule, nil, nil, nil)

	resp, err := srv.CreateInvoiceFromSchedule(context.Background(), &invoiceGrpc.CreateInvoiceFromScheduleRequest{
		UserId: "user-1", ScheduleId: "sched-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success || len(resp.ValidationErrors) == 0 {
		t.Fatal("expected a validation error when no months are selected")
	}
}
