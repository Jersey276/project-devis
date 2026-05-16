package tests

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"project-devis-quote/actions"
	quoteGrpc "project-devis-quote/services/grpc"
)

func TestCreateQuote_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`INSERT INTO quotes`).
		WithArgs(sqlmock.AnyArg(), "user-1", "Devis Test", "client-1", int32(7)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.CreateQuote(context.Background(), &quoteGrpc.CreateQuoteRequest{
		UserId:    "user-1",
		Name:      "Devis Test",
		ClientId:  "client-1",
		AddressId: 7,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.QuoteId == "" {
		t.Fatal("expected non-empty quote_id")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateQuote_MissingName(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateQuote(context.Background(), &quoteGrpc.CreateQuoteRequest{
		UserId:    "user-1",
		ClientId:  "client-1",
		AddressId: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for missing name")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestCreateQuote_MissingUser(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.CreateQuote(context.Background(), &quoteGrpc.CreateQuoteRequest{
		Name:      "x",
		ClientId:  "client-1",
		AddressId: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for missing user_id")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
}

func TestCreateQuote_MissingClient(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.CreateQuote(context.Background(), &quoteGrpc.CreateQuoteRequest{
		UserId:    "user-1",
		Name:      "x",
		AddressId: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
}

func TestCreateQuote_MissingAddress(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.CreateQuote(context.Background(), &quoteGrpc.CreateQuoteRequest{
		UserId:   "user-1",
		Name:     "x",
		ClientId: "client-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
}

func TestGetQuote_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT quote_id, user_id, name, archived_at, state, client_id, address_id FROM quotes`).
		WithArgs("q-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id", "user_id", "name", "archived_at", "state", "client_id", "address_id"}).
			AddRow("q-1", "user-1", "Devis", nil, "draft", "client-1", int32(3)))

	mock.ExpectQuery(`SELECT line_id, quote_id, type, name`).
		WithArgs("q-1").
		WillReturnRows(sqlmock.NewRows([]string{"line_id", "quote_id", "type", "name", "quantity", "unit", "unit_price", "data", "position", "tax_id"}).
			AddRow("l-1", "q-1", "simple", "Item", "2", "u", int64(1500), "{}", int32(0), int32(0)))

	resp, err := srv.GetQuote(context.Background(), &quoteGrpc.GetQuoteRequest{
		QuoteId: "q-1",
		UserId:  "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Quote.QuoteId != "q-1" || resp.Quote.Name != "Devis" {
		t.Fatalf("unexpected quote payload: %+v", resp.Quote)
	}
	if resp.Quote.Archived {
		t.Fatal("expected archived=false when archived_at is NULL")
	}
	if len(resp.Lines) != 1 || resp.Lines[0].LineId != "l-1" {
		t.Fatalf("expected 1 line, got %+v", resp.Lines)
	}
}

func TestGetQuote_Archived(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT quote_id, user_id, name, archived_at, state, client_id, address_id FROM quotes`).
		WithArgs("q-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id", "user_id", "name", "archived_at", "state", "client_id", "address_id"}).
			AddRow("q-1", "user-1", "Devis", time.Now(), "draft", "client-1", int32(3)))

	mock.ExpectQuery(`SELECT line_id, quote_id, type, name`).
		WithArgs("q-1").
		WillReturnRows(sqlmock.NewRows([]string{"line_id", "quote_id", "type", "name", "quantity", "unit", "unit_price", "data", "position"}))

	resp, err := srv.GetQuote(context.Background(), &quoteGrpc.GetQuoteRequest{
		QuoteId: "q-1",
		UserId:  "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if !resp.Quote.Archived {
		t.Fatal("expected archived=true when archived_at is set")
	}
}

func TestGetQuote_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT quote_id, user_id, name, archived_at, state, client_id, address_id FROM quotes`).
		WithArgs("ghost", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id", "user_id", "name", "archived_at", "state", "client_id", "address_id"}))

	resp, err := srv.GetQuote(context.Background(), &quoteGrpc.GetQuoteRequest{
		QuoteId: "ghost",
		UserId:  "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for unknown quote")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
}

func TestListQuotes_ExcludesArchivedByDefault(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT quote_id, user_id, name, archived_at, state, client_id, address_id FROM quotes WHERE user_id=\$1 AND archived_at IS NULL`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id", "user_id", "name", "archived_at", "state", "client_id", "address_id"}).
			AddRow("q-1", "user-1", "A", nil, "draft", "client-1", int32(1)).
			AddRow("q-2", "user-1", "B", nil, "draft", "client-1", int32(1)))

	resp, err := srv.ListQuotes(context.Background(), &quoteGrpc.ListQuotesRequest{UserId: "user-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if len(resp.Quotes) != 2 {
		t.Fatalf("expected 2 quotes, got %d", len(resp.Quotes))
	}
}

func TestListQuotes_IncludeArchived(t *testing.T) {
	srv, mock := setupServer(t)

	// When include_archived=true, the WHERE clause must NOT contain archived_at IS NULL
	mock.ExpectQuery(`SELECT quote_id, user_id, name, archived_at, state, client_id, address_id FROM quotes WHERE user_id=\$1 ORDER`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id", "user_id", "name", "archived_at", "state", "client_id", "address_id"}).
			AddRow("q-1", "user-1", "A", nil, "draft", "client-1", int32(1)).
			AddRow("q-2", "user-1", "B", time.Now(), "draft", "client-1", int32(1)))

	resp, err := srv.ListQuotes(context.Background(), &quoteGrpc.ListQuotesRequest{
		UserId:          "user-1",
		IncludeArchived: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if len(resp.Quotes) != 2 {
		t.Fatalf("expected 2 quotes, got %d", len(resp.Quotes))
	}
	if !resp.Quotes[1].Archived {
		t.Fatal("expected second quote to be archived")
	}
}

func TestUpdateQuote_Success(t *testing.T) {
	srv, mock := setupServer(t)

	expectEditableCheck(mock, "q-1", "user-1", "draft")
	// Name-only PUT: empty ClientId / zero AddressId are the "preserve" sentinels.
	mock.ExpectExec(`UPDATE quotes SET\s+name`).
		WithArgs("q-1", "user-1", "New name", "", int32(0)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.UpdateQuote(context.Background(), &quoteGrpc.UpdateQuoteRequest{
		QuoteId: "q-1",
		UserId:  "user-1",
		Name:    "New name",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestUpdateQuote_ChangesPickers asserts the COALESCE(NULLIF) preserve-pattern:
// non-empty ClientId / non-zero AddressId flow through unchanged so the SQL
// overwrites the existing values.
func TestUpdateQuote_ChangesPickers(t *testing.T) {
	srv, mock := setupServer(t)

	expectEditableCheck(mock, "q-1", "user-1", "draft")
	mock.ExpectExec(`UPDATE quotes SET\s+name`).
		WithArgs("q-1", "user-1", "Renamed", "c-7", int32(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.UpdateQuote(context.Background(), &quoteGrpc.UpdateQuoteRequest{
		QuoteId:   "q-1",
		UserId:    "user-1",
		Name:      "Renamed",
		ClientId:  "c-7",
		AddressId: 42,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateQuote_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	expectEditableCheck(mock, "q-1", "user-1", "")

	resp, err := srv.UpdateQuote(context.Background(), &quoteGrpc.UpdateQuoteRequest{
		QuoteId: "q-1",
		UserId:  "user-1",
		Name:    "x",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure when quote does not exist")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
}

func TestUpdateQuote_BlockedWhenValidated(t *testing.T) {
	srv, mock := setupServer(t)

	expectEditableCheck(mock, "q-1", "user-1", "validated")

	resp, err := srv.UpdateQuote(context.Background(), &quoteGrpc.UpdateQuoteRequest{
		QuoteId: "q-1",
		UserId:  "user-1",
		Name:    "x",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != actions.CodeQuoteFinalized {
		t.Fatalf("expected CodeQuoteFinalized, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestUpdateQuote_BlockedWhenDrop(t *testing.T) {
	srv, mock := setupServer(t)

	expectEditableCheck(mock, "q-1", "user-1", "drop")

	resp, err := srv.UpdateQuote(context.Background(), &quoteGrpc.UpdateQuoteRequest{
		QuoteId: "q-1",
		UserId:  "user-1",
		Name:    "x",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != actions.CodeQuoteFinalized {
		t.Fatalf("expected CodeQuoteFinalized, got %d", resp.Code)
	}
}

func TestUpdateQuote_AllowedWhenSent(t *testing.T) {
	srv, mock := setupServer(t)

	expectEditableCheck(mock, "q-1", "user-1", "sent")
	mock.ExpectExec(`UPDATE quotes SET\s+name`).
		WithArgs("q-1", "user-1", "Updated", "", int32(0)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.UpdateQuote(context.Background(), &quoteGrpc.UpdateQuoteRequest{
		QuoteId: "q-1",
		UserId:  "user-1",
		Name:    "Updated",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

func TestDropQuote_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE quotes SET state='drop'`).
		WithArgs("q-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.DropQuote(context.Background(), &quoteGrpc.DropQuoteRequest{
		QuoteId: "q-1",
		UserId:  "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

func TestDropQuote_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE quotes SET state='drop'`).
		WithArgs("q-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, err := srv.DropQuote(context.Background(), &quoteGrpc.DropQuoteRequest{
		QuoteId: "q-1",
		UserId:  "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
}

func TestDropQuote_MissingInput(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.DropQuote(context.Background(), &quoteGrpc.DropQuoteRequest{UserId: "user-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
}

func TestContinueQuote_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE quotes SET state='draft'`).
		WithArgs("q-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.ContinueQuote(context.Background(), &quoteGrpc.ContinueQuoteRequest{
		QuoteId: "q-1",
		UserId:  "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

func TestContinueQuote_NotInDropState(t *testing.T) {
	srv, mock := setupServer(t)

	// rows=0 because the WHERE state='drop' filter excluded it
	mock.ExpectExec(`UPDATE quotes SET state='draft'`).
		WithArgs("q-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, err := srv.ContinueQuote(context.Background(), &quoteGrpc.ContinueQuoteRequest{
		QuoteId: "q-1",
		UserId:  "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
}

func TestDeleteQuote_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`DELETE FROM quotes`).
		WithArgs("q-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.DeleteQuote(context.Background(), &quoteGrpc.DeleteQuoteRequest{
		QuoteId: "q-1",
		UserId:  "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

func TestDeleteQuote_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`DELETE FROM quotes`).
		WithArgs("q-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, err := srv.DeleteQuote(context.Background(), &quoteGrpc.DeleteQuoteRequest{
		QuoteId: "q-1",
		UserId:  "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
}

func TestArchiveQuote_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE quotes SET archived_at=NOW`).
		WithArgs("q-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.ArchiveQuote(context.Background(), &quoteGrpc.ArchiveQuoteRequest{
		QuoteId: "q-1",
		UserId:  "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

func TestArchiveQuote_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE quotes SET archived_at=NOW`).
		WithArgs("q-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, err := srv.ArchiveQuote(context.Background(), &quoteGrpc.ArchiveQuoteRequest{
		QuoteId: "q-1",
		UserId:  "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
}

func TestRestoreQuote_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE quotes SET archived_at=NULL`).
		WithArgs("q-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.RestoreQuote(context.Background(), &quoteGrpc.RestoreQuoteRequest{
		QuoteId: "q-1",
		UserId:  "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

func TestTrashQuotes_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`DELETE FROM quotes WHERE user_id=\$1 AND archived_at IS NOT NULL`).
		WithArgs("user-1").
		WillReturnResult(sqlmock.NewResult(0, 3))

	resp, err := srv.TrashQuotes(context.Background(), &quoteGrpc.TrashQuotesRequest{UserId: "user-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

func TestTrashQuotes_MissingUser(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.TrashQuotes(context.Background(), &quoteGrpc.TrashQuotesRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
}
