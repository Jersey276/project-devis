package tests

import (
	"context"
	"testing"

	"project-devis-quote/actions"
	quoteGrpc "project-devis-quote/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

const (
	multipleData = `{"kind":"detailed","sublines":[{"name":"a","quantity":"1","unit_price":1000}]}`
)

func TestCreateLine_SimpleSuccess(t *testing.T) {
	srv, mock := setupServer(t)

	expectEditableCheck(mock, "q-1", "user-1", "draft")
	mock.ExpectExec(`INSERT INTO quote_lines`).
		WithArgs(sqlmock.AnyArg(), "q-1", "simple", "Item", "2", "u", int64(1500), `{"kind":"line"}`, int32(0), nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.CreateQuoteLine(context.Background(), &quoteGrpc.CreateQuoteLineRequest{
		QuoteId:   "q-1",
		UserId:    "user-1",
		Type:      "simple",
		Name:      "Item",
		Quantity:  "2",
		Unit:      "u",
		UnitPrice: 1500,
		Data:      `{"kind":"line"}`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.LineId == "" {
		t.Fatal("expected non-empty line_id")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateLine_MultipleSuccess(t *testing.T) {
	srv, mock := setupServer(t)

	expectEditableCheck(mock, "q-1", "user-1", "draft")
	mock.ExpectExec(`INSERT INTO quote_lines`).
		WithArgs(sqlmock.AnyArg(), "q-1", "multiple", "Pack", "1", nil, int64(0), multipleData, int32(0), nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.CreateQuoteLine(context.Background(), &quoteGrpc.CreateQuoteLineRequest{
		QuoteId:  "q-1",
		UserId:   "user-1",
		Type:     "multiple",
		Name:     "Pack",
		Quantity: "1",
		Data:     multipleData,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

func TestCreateLine_InvalidType(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateQuoteLine(context.Background(), &quoteGrpc.CreateQuoteLineRequest{
		QuoteId:  "q-1",
		UserId:   "user-1",
		Type:     "weird",
		Name:     "X",
		Quantity: "1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for unknown type")
	}
	if resp.Code != actions.CodeInvalidLineType {
		t.Fatalf("expected CodeInvalidLineType, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestCreateLine_InvalidData_MultipleNoSublines(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateQuoteLine(context.Background(), &quoteGrpc.CreateQuoteLineRequest{
		QuoteId:  "q-1",
		UserId:   "user-1",
		Type:     "multiple",
		Name:     "X",
		Quantity: "1",
		Data:     `{"sublines":[]}`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for empty sublines")
	}
	if resp.Code != actions.CodeInvalidLineData {
		t.Fatalf("expected CodeInvalidLineData, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestCreateLine_InvalidData_SimpleNonEmpty(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.CreateQuoteLine(context.Background(), &quoteGrpc.CreateQuoteLineRequest{
		QuoteId:  "q-1",
		UserId:   "user-1",
		Type:     "simple",
		Name:     "X",
		Quantity: "1",
		Data:     `{"foo":"bar"}`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != actions.CodeInvalidLineData {
		t.Fatalf("expected CodeInvalidLineData, got %d", resp.Code)
	}
}

func TestCreateLine_OwnerNotFound(t *testing.T) {
	srv, mock := setupServer(t)

	expectEditableCheck(mock, "q-1", "user-1", "")

	resp, err := srv.CreateQuoteLine(context.Background(), &quoteGrpc.CreateQuoteLineRequest{
		QuoteId:  "q-1",
		UserId:   "user-1",
		Type:     "simple",
		Name:     "X",
		Quantity: "1",
		Data:     `{"kind":"line"}`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure when user does not own quote")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
}

func TestCreateLine_BlockedWhenFinalized(t *testing.T) {
	srv, mock := setupServer(t)

	expectEditableCheck(mock, "q-1", "user-1", "validated")

	resp, err := srv.CreateQuoteLine(context.Background(), &quoteGrpc.CreateQuoteLineRequest{
		QuoteId:  "q-1",
		UserId:   "user-1",
		Type:     "simple",
		Name:     "X",
		Quantity: "1",
		Data:     `{"kind":"line"}`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != actions.CodeQuoteFinalized {
		t.Fatalf("expected CodeQuoteFinalized, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateLine_NegativeUnitPrice(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.CreateQuoteLine(context.Background(), &quoteGrpc.CreateQuoteLineRequest{
		QuoteId:   "q-1",
		UserId:    "user-1",
		Type:      "simple",
		Name:      "X",
		Quantity:  "1",
		UnitPrice: -100,
		Data:      `{"kind":"line"}`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
}

func TestCreateLine_NonNumericQuantity(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.CreateQuoteLine(context.Background(), &quoteGrpc.CreateQuoteLineRequest{
		QuoteId:  "q-1",
		UserId:   "user-1",
		Type:     "simple",
		Name:     "X",
		Quantity: "abc",
		Data:     `{"kind":"line"}`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
}

func TestGetLine_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT l.line_id, l.quote_id`).
		WithArgs("l-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"line_id", "quote_id", "type", "name", "quantity", "unit", "unit_price", "data", "position", "tax_id"}).
			AddRow("l-1", "q-1", "simple", "Item", "2", "u", int64(1500), `{"kind":"line"}`, int32(0), int32(0)))

	resp, err := srv.GetQuoteLine(context.Background(), &quoteGrpc.GetQuoteLineRequest{
		LineId: "l-1",
		UserId: "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Line.LineId != "l-1" || resp.Line.UnitPrice != 1500 {
		t.Fatalf("unexpected line payload: %+v", resp.Line)
	}
}

func TestGetLine_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT l.line_id, l.quote_id`).
		WithArgs("ghost", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"line_id", "quote_id", "type", "name", "quantity", "unit", "unit_price", "data", "position", "tax_id"}))

	resp, err := srv.GetQuoteLine(context.Background(), &quoteGrpc.GetQuoteLineRequest{
		LineId: "ghost",
		UserId: "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for unknown line")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
}

func TestListLines_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM quotes WHERE quote_id=\$1 AND user_id=\$2\)`).
		WithArgs("q-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery(`SELECT line_id, quote_id, type, name`).
		WithArgs("q-1").
		WillReturnRows(sqlmock.NewRows([]string{"line_id", "quote_id", "type", "name", "quantity", "unit", "unit_price", "data", "position", "tax_id"}).
			AddRow("l-1", "q-1", "simple", "A", "1", "", int64(100), `{"kind":"line"}`, int32(0), int32(0)).
			AddRow("l-2", "q-1", "multiple", "B", "1", "", int64(0), multipleData, int32(1), int32(0)))

	resp, err := srv.ListQuoteLines(context.Background(), &quoteGrpc.ListQuoteLinesRequest{
		QuoteId: "q-1",
		UserId:  "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if len(resp.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(resp.Lines))
	}
}

func TestListLines_OwnerNotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM quotes WHERE quote_id=\$1 AND user_id=\$2\)`).
		WithArgs("q-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	resp, err := srv.ListQuoteLines(context.Background(), &quoteGrpc.ListQuoteLinesRequest{
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

func TestUpdateLine_Success(t *testing.T) {
	srv, mock := setupServer(t)

	expectLineParentEditable(mock, "l-1", "user-1", "draft")
	mock.ExpectExec(`UPDATE quote_lines\s+SET type`).
		WithArgs("simple", "Item v2", "3", "u", int64(2000), `{"kind":"line"}`, int32(2), nil, "l-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.UpdateQuoteLine(context.Background(), &quoteGrpc.UpdateQuoteLineRequest{
		LineId:    "l-1",
		UserId:    "user-1",
		Type:      "simple",
		Name:      "Item v2",
		Quantity:  "3",
		Unit:      "u",
		UnitPrice: 2000,
		Data:      `{"kind":"line"}`,
		Position:  2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

func TestUpdateLine_LineNotFound(t *testing.T) {
	srv, mock := setupServer(t)

	expectLineParentEditable(mock, "l-1", "user-1", "")

	resp, err := srv.UpdateQuoteLine(context.Background(), &quoteGrpc.UpdateQuoteLineRequest{
		LineId:   "l-1",
		UserId:   "user-1",
		Type:     "simple",
		Name:     "X",
		Quantity: "1",
		Data:     `{"kind":"line"}`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
}

func TestUpdateLine_BlockedWhenFinalized(t *testing.T) {
	srv, mock := setupServer(t)

	expectLineParentEditable(mock, "l-1", "user-1", "drop")

	resp, err := srv.UpdateQuoteLine(context.Background(), &quoteGrpc.UpdateQuoteLineRequest{
		LineId:   "l-1",
		UserId:   "user-1",
		Type:     "simple",
		Name:     "X",
		Quantity: "1",
		Data:     `{"kind":"line"}`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != actions.CodeQuoteFinalized {
		t.Fatalf("expected CodeQuoteFinalized, got %d", resp.Code)
	}
}

func TestDeleteLine_Success(t *testing.T) {
	srv, mock := setupServer(t)

	expectLineParentEditable(mock, "l-1", "user-1", "draft")
	mock.ExpectExec(`DELETE FROM quote_lines WHERE line_id=\$1`).
		WithArgs("l-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.DeleteQuoteLine(context.Background(), &quoteGrpc.DeleteQuoteLineRequest{
		LineId: "l-1",
		UserId: "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

func TestDeleteLine_LineNotFound(t *testing.T) {
	srv, mock := setupServer(t)

	expectLineParentEditable(mock, "l-1", "user-1", "")

	resp, err := srv.DeleteQuoteLine(context.Background(), &quoteGrpc.DeleteQuoteLineRequest{
		LineId: "l-1",
		UserId: "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
}

func TestDeleteLine_BlockedWhenFinalized(t *testing.T) {
	srv, mock := setupServer(t)

	expectLineParentEditable(mock, "l-1", "user-1", "validated")

	resp, err := srv.DeleteQuoteLine(context.Background(), &quoteGrpc.DeleteQuoteLineRequest{
		LineId: "l-1",
		UserId: "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != actions.CodeQuoteFinalized {
		t.Fatalf("expected CodeQuoteFinalized, got %d", resp.Code)
	}
}
