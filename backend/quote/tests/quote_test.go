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
		WithArgs(sqlmock.AnyArg(), "user-1", "Devis Test").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.CreateQuote(context.Background(), &quoteGrpc.CreateQuoteRequest{
		UserId: "user-1",
		Name:   "Devis Test",
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
		UserId: "user-1",
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

	resp, err := srv.CreateQuote(context.Background(), &quoteGrpc.CreateQuoteRequest{Name: "x"})
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

func TestGetQuote_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT quote_id, user_id, name, archived_at FROM quotes`).
		WithArgs("q-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id", "user_id", "name", "archived_at"}).
			AddRow("q-1", "user-1", "Devis", nil))

	mock.ExpectQuery(`SELECT line_id, quote_id, type, name`).
		WithArgs("q-1").
		WillReturnRows(sqlmock.NewRows([]string{"line_id", "quote_id", "type", "name", "quantity", "unit", "unit_price", "data", "position"}).
			AddRow("l-1", "q-1", "simple", "Item", "2", "u", int64(1500), "{}", int32(0)))

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

	mock.ExpectQuery(`SELECT quote_id, user_id, name, archived_at FROM quotes`).
		WithArgs("q-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id", "user_id", "name", "archived_at"}).
			AddRow("q-1", "user-1", "Devis", time.Now()))

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

	mock.ExpectQuery(`SELECT quote_id, user_id, name, archived_at FROM quotes`).
		WithArgs("ghost", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id", "user_id", "name", "archived_at"}))

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

	mock.ExpectQuery(`SELECT quote_id, user_id, name, archived_at FROM quotes WHERE user_id=\$1 AND archived_at IS NULL`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id", "user_id", "name", "archived_at"}).
			AddRow("q-1", "user-1", "A", nil).
			AddRow("q-2", "user-1", "B", nil))

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
	mock.ExpectQuery(`SELECT quote_id, user_id, name, archived_at FROM quotes WHERE user_id=\$1 ORDER`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id", "user_id", "name", "archived_at"}).
			AddRow("q-1", "user-1", "A", nil).
			AddRow("q-2", "user-1", "B", time.Now()))

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

	mock.ExpectExec(`UPDATE quotes SET name`).
		WithArgs("New name", "q-1", "user-1").
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
}

func TestUpdateQuote_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE quotes SET name`).
		WithArgs("x", "q-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, err := srv.UpdateQuote(context.Background(), &quoteGrpc.UpdateQuoteRequest{
		QuoteId: "q-1",
		UserId:  "user-1",
		Name:    "x",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure when no rows affected")
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
