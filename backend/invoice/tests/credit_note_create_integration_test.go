package tests

import (
	"context"
	"testing"

	"google.golang.org/grpc"

	"project-devis-invoice/actions"
	invoiceGrpc "project-devis-invoice/services/grpc"
	quoteGrpc "project-devis-invoice/services/quotegrpc"
)

// issueTwoLineInvoice creates and issues an invoice with two lines (20000 +
// 10000 HT, both at 20% VAT) so credit note tests can exercise partial vs
// total crediting.
func issueTwoLineInvoice(t *testing.T, srv *actions.Server, userID, quoteID string) string {
	t.Helper()
	resp, err := srv.CreateInvoiceFromQuote(context.Background(), &invoiceGrpc.CreateInvoiceFromQuoteRequest{
		UserId: userID, QuoteId: quoteID, IssueNow: true,
	})
	if err != nil {
		t.Fatalf("create+issue invoice: %v", err)
	}
	if !resp.Success {
		t.Fatalf("create+issue invoice failed: code=%d", resp.Code)
	}
	return resp.InvoiceId
}

func twoLineQuoteFixtures() (*mockQuoteClient, *mockScheduleClient, *mockUsersClient) {
	quote, schedule, users := domesticFixtures()
	quote.GetQuoteFn = func(_ context.Context, req *quoteGrpc.GetQuoteRequest, _ ...grpc.CallOption) (*quoteGrpc.GetQuoteResponse, error) {
		return &quoteGrpc.GetQuoteResponse{
			Success: true,
			Quote: &quoteGrpc.Quote{
				QuoteId: req.QuoteId, UserId: req.UserId, State: quoteGrpc.QuoteState_QUOTE_STATE_VALIDATED,
				ClientId: "client-1", AddressId: 10, UserAddressId: 20,
			},
			Lines: []*quoteGrpc.QuoteLine{
				{LineId: "line-1", QuoteId: req.QuoteId, Name: "Prestation A", Unit: "u", Quantity: "1", UnitPrice: 20000, Position: 0, TaxId: 1},
				{LineId: "line-2", QuoteId: req.QuoteId, Name: "Prestation B", Unit: "u", Quantity: "1", UnitPrice: 10000, Position: 1, TaxId: 1},
			},
		}, nil
	}
	return quote, schedule, users
}

func TestCreateCreditNote_PartialCredit_OneOfTwoLines(t *testing.T) {
	db := sealTestDB(t)
	quote, schedule, users := twoLineQuoteFixtures()
	srv := actions.NewServer(db, quote, users, schedule, nil, nil, nil)
	invoiceID := issueTwoLineInvoice(t, srv, "user-1", "quote-1")

	resp, err := srv.CreateCreditNote(context.Background(), &invoiceGrpc.CreateCreditNoteRequest{
		UserId: "user-1", InvoiceId: invoiceID, Positions: []int32{0}, Reason: "Ligne annulée",
	})
	if err != nil {
		t.Fatalf("create credit note: %v", err)
	}
	if !resp.Success || resp.CreditNoteNumber == "" {
		t.Fatalf("create credit note failed: success=%v code=%d", resp.Success, resp.Code)
	}

	get, err := srv.GetCreditNote(context.Background(), &invoiceGrpc.GetCreditNoteRequest{CreditNoteId: resp.CreditNoteId, UserId: "user-1"})
	if err != nil {
		t.Fatalf("get credit note: %v", err)
	}
	if !get.Success {
		t.Fatalf("get credit note failed: code=%d", get.Code)
	}
	if got := get.CreditNote.GetIsTotal(); got {
		t.Error("crediting only 1 of 2 lines should not be marked as total")
	}
	if got := get.CreditNote.GetTotalHtCents(); got != 20000 {
		t.Errorf("credit note HT = %d; want 20000 (line 0 only)", got)
	}
	if got := get.CreditNote.GetTotalTtcCents(); got != 24000 {
		t.Errorf("credit note TTC = %d; want 24000", got)
	}
	if len(get.CreditNote.GetLines()) != 1 {
		t.Fatalf("expected 1 credited line, got %d", len(get.CreditNote.GetLines()))
	}
}

func TestCreateCreditNote_NoPositions_CreditsAllRemaining_MarksTotal(t *testing.T) {
	db := sealTestDB(t)
	quote, schedule, users := twoLineQuoteFixtures()
	srv := actions.NewServer(db, quote, users, schedule, nil, nil, nil)
	invoiceID := issueTwoLineInvoice(t, srv, "user-1", "quote-1")

	resp, err := srv.CreateCreditNote(context.Background(), &invoiceGrpc.CreateCreditNoteRequest{
		UserId: "user-1", InvoiceId: invoiceID,
	})
	if err != nil {
		t.Fatalf("create credit note: %v", err)
	}
	if !resp.Success {
		t.Fatalf("create credit note failed: code=%d", resp.Code)
	}

	get, err := srv.GetCreditNote(context.Background(), &invoiceGrpc.GetCreditNoteRequest{CreditNoteId: resp.CreditNoteId, UserId: "user-1"})
	if err != nil {
		t.Fatalf("get credit note: %v", err)
	}
	if !get.CreditNote.GetIsTotal() {
		t.Error("crediting all remaining lines should be marked as total")
	}
	if got := get.CreditNote.GetTotalHtCents(); got != 30000 {
		t.Errorf("credit note HT = %d; want 30000 (both lines)", got)
	}
}

func TestCreateCreditNote_SameLineTwice_Refused(t *testing.T) {
	db := sealTestDB(t)
	quote, schedule, users := twoLineQuoteFixtures()
	srv := actions.NewServer(db, quote, users, schedule, nil, nil, nil)
	invoiceID := issueTwoLineInvoice(t, srv, "user-1", "quote-1")

	first, err := srv.CreateCreditNote(context.Background(), &invoiceGrpc.CreateCreditNoteRequest{
		UserId: "user-1", InvoiceId: invoiceID, Positions: []int32{0},
	})
	if err != nil || !first.Success {
		t.Fatalf("first credit note failed: err=%v code=%d", err, first.Code)
	}

	second, err := srv.CreateCreditNote(context.Background(), &invoiceGrpc.CreateCreditNoteRequest{
		UserId: "user-1", InvoiceId: invoiceID, Positions: []int32{0},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if second.Success {
		t.Fatal("expected refusal when crediting an already-credited line")
	}
}

func TestCreateCreditNote_DraftInvoice_Refused(t *testing.T) {
	db := sealTestDB(t)
	quote, schedule, users := twoLineQuoteFixtures()
	srv := actions.NewServer(db, quote, users, schedule, nil, nil, nil)

	created, err := srv.CreateInvoiceFromQuote(context.Background(), &invoiceGrpc.CreateInvoiceFromQuoteRequest{
		UserId: "user-1", QuoteId: "quote-1",
	})
	if err != nil || !created.Success {
		t.Fatalf("create draft invoice failed: err=%v code=%d", err, created.Code)
	}

	resp, err := srv.CreateCreditNote(context.Background(), &invoiceGrpc.CreateCreditNoteRequest{
		UserId: "user-1", InvoiceId: created.InvoiceId,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected refusal when crediting a DRAFT invoice")
	}
}

func TestListCreditNotes_ReturnsCreated(t *testing.T) {
	db := sealTestDB(t)
	quote, schedule, users := twoLineQuoteFixtures()
	srv := actions.NewServer(db, quote, users, schedule, nil, nil, nil)
	invoiceID := issueTwoLineInvoice(t, srv, "user-1", "quote-1")

	if _, err := srv.CreateCreditNote(context.Background(), &invoiceGrpc.CreateCreditNoteRequest{
		UserId: "user-1", InvoiceId: invoiceID, Positions: []int32{0},
	}); err != nil {
		t.Fatalf("create credit note: %v", err)
	}

	list, err := srv.ListCreditNotes(context.Background(), &invoiceGrpc.ListCreditNotesRequest{UserId: "user-1"})
	if err != nil {
		t.Fatalf("list credit notes: %v", err)
	}
	if !list.Success || list.Total != 1 || len(list.CreditNotes) != 1 {
		t.Fatalf("expected 1 credit note, got total=%d len=%d", list.Total, len(list.CreditNotes))
	}
	if list.CreditNotes[0].GetInvoiceId() != invoiceID {
		t.Errorf("credit note invoice_id = %q; want %q", list.CreditNotes[0].GetInvoiceId(), invoiceID)
	}
}

func TestListInvoices_FiltersByStatus(t *testing.T) {
	db := sealTestDB(t)
	quote, schedule, users := twoLineQuoteFixtures()
	srv := actions.NewServer(db, quote, users, schedule, nil, nil, nil)

	issueTwoLineInvoice(t, srv, "user-1", "quote-1")
	if _, err := srv.CreateInvoiceFromQuote(context.Background(), &invoiceGrpc.CreateInvoiceFromQuoteRequest{
		UserId: "user-1", QuoteId: "quote-2",
	}); err != nil {
		t.Fatalf("create draft invoice: %v", err)
	}

	all, err := srv.ListInvoices(context.Background(), &invoiceGrpc.ListInvoicesRequest{UserId: "user-1"})
	if err != nil {
		t.Fatalf("list invoices: %v", err)
	}
	if all.Total != 2 {
		t.Fatalf("expected 2 invoices total, got %d", all.Total)
	}

	issued, err := srv.ListInvoices(context.Background(), &invoiceGrpc.ListInvoicesRequest{
		UserId: "user-1", Filters: &invoiceGrpc.InvoiceFilters{Statuses: []string{"ISSUED"}},
	})
	if err != nil {
		t.Fatalf("list issued invoices: %v", err)
	}
	if issued.Total != 1 || len(issued.Invoices) != 1 {
		t.Fatalf("expected 1 ISSUED invoice, got total=%d len=%d", issued.Total, len(issued.Invoices))
	}
	if issued.Invoices[0].GetStatus() != "ISSUED" {
		t.Errorf("status = %q; want ISSUED", issued.Invoices[0].GetStatus())
	}
}
