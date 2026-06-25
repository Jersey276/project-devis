package tests

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"project-devis-quote/actions"
	quoteGrpc "project-devis-quote/services/grpc"
)

var commentCols = []string{
	"comment_id", "line_id", "quote_id", "author_id", "author_name", "body", "created_at", "updated_at",
}

func commentRow(id, lineID, quoteID, authorID, authorName, body string) *sqlmock.Rows {
	ts := time.Now().Format(time.RFC3339)
	return sqlmock.NewRows(commentCols).AddRow(id, lineID, quoteID, authorID, authorName, body, ts, ts)
}

// ─── Create ──────────────────────────────────────────────────────────────────

func TestCreateComment_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`INSERT INTO quote_line_comments`).
		WithArgs("line-1", "quote-1", "user-1", "Alice", "Mon commentaire").
		WillReturnRows(commentRow("cmt-1", "line-1", "quote-1", "user-1", "Alice", "Mon commentaire"))

	resp, err := srv.CreateComment(context.Background(), &quoteGrpc.CreateCommentRequest{
		LineId:     "line-1",
		QuoteId:    "quote-1",
		AuthorId:   "user-1",
		AuthorName: "Alice",
		Body:       "Mon commentaire",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Comment == nil || resp.Comment.CommentId != "cmt-1" {
		t.Fatal("expected comment in response")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// AuthorName vide → doit être remplacé par "Inconnu".
func TestCreateComment_DefaultAuthorName(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`INSERT INTO quote_line_comments`).
		WithArgs("line-1", "quote-1", "user-1", "Inconnu", "body").
		WillReturnRows(commentRow("cmt-2", "line-1", "quote-1", "user-1", "Inconnu", "body"))

	resp, err := srv.CreateComment(context.Background(), &quoteGrpc.CreateCommentRequest{
		LineId:   "line-1",
		QuoteId:  "quote-1",
		AuthorId: "user-1",
		Body:     "body",
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

func TestCreateComment_MissingLineId(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateComment(context.Background(), &quoteGrpc.CreateCommentRequest{
		QuoteId:  "quote-1",
		AuthorId: "user-1",
		Body:     "body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for missing line_id")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestCreateComment_MissingBody(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateComment(context.Background(), &quoteGrpc.CreateCommentRequest{
		LineId:   "line-1",
		QuoteId:  "quote-1",
		AuthorId: "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for missing body")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

// ─── List ─────────────────────────────────────────────────────────────────────

func TestListComments_Success(t *testing.T) {
	srv, mock := setupServer(t)

	rows := commentRow("cmt-1", "line-1", "quote-1", "user-1", "Alice", "Premier")
	rows.AddRow("cmt-2", "line-1", "quote-1", "user-2", "Bob", "Deuxième",
		time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))

	mock.ExpectQuery(`SELECT .+ FROM quote_line_comments WHERE line_id`).
		WithArgs("line-1").
		WillReturnRows(rows)

	resp, err := srv.ListComments(context.Background(), &quoteGrpc.ListCommentsRequest{
		LineId: "line-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if len(resp.Comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(resp.Comments))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListComments_Empty(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT .+ FROM quote_line_comments WHERE line_id`).
		WithArgs("line-empty").
		WillReturnRows(sqlmock.NewRows(commentCols))

	resp, err := srv.ListComments(context.Background(), &quoteGrpc.ListCommentsRequest{
		LineId: "line-empty",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success for empty list, got code %d", resp.Code)
	}
	if len(resp.Comments) != 0 {
		t.Fatalf("expected empty list, got %d comments", len(resp.Comments))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListComments_MissingLineId(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.ListComments(context.Background(), &quoteGrpc.ListCommentsRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for missing line_id")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

// ─── Update ───────────────────────────────────────────────────────────────────

func TestUpdateComment_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`UPDATE quote_line_comments`).
		WithArgs("nouveau corps", "cmt-1", "user-1").
		WillReturnRows(commentRow("cmt-1", "line-1", "quote-1", "user-1", "Alice", "nouveau corps"))

	resp, err := srv.UpdateComment(context.Background(), &quoteGrpc.UpdateCommentRequest{
		CommentId: "cmt-1",
		AuthorId:  "user-1",
		Body:      "nouveau corps",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Comment == nil || resp.Comment.Body != "nouveau corps" {
		t.Fatal("expected updated comment in response")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// L'auteur ne correspond pas → UPDATE retourne 0 ligne, le commentaire existe → CommentForbidden.
func TestUpdateComment_Forbidden(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`UPDATE quote_line_comments`).
		WithArgs("body", "cmt-1", "autre-user").
		WillReturnRows(sqlmock.NewRows(commentCols)) // 0 ligne

	// commentExists : le commentaire existe bien.
	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("cmt-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	resp, err := srv.UpdateComment(context.Background(), &quoteGrpc.UpdateCommentRequest{
		CommentId: "cmt-1",
		AuthorId:  "autre-user",
		Body:      "body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for wrong author")
	}
	if resp.Code != actions.CodeCommentForbidden {
		t.Fatalf("expected CodeCommentForbidden, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// UPDATE 0 ligne et le commentaire n'existe pas → NotFound.
func TestUpdateComment_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`UPDATE quote_line_comments`).
		WithArgs("body", "cmt-inexistant", "user-1").
		WillReturnRows(sqlmock.NewRows(commentCols))

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("cmt-inexistant").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	resp, err := srv.UpdateComment(context.Background(), &quoteGrpc.UpdateCommentRequest{
		CommentId: "cmt-inexistant",
		AuthorId:  "user-1",
		Body:      "body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for non-existent comment")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateComment_MissingFields(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.UpdateComment(context.Background(), &quoteGrpc.UpdateCommentRequest{
		CommentId: "cmt-1",
		// AuthorId et Body manquants
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for missing fields")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func TestDeleteComment_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`DELETE FROM quote_line_comments`).
		WithArgs("cmt-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.DeleteComment(context.Background(), &quoteGrpc.DeleteCommentRequest{
		CommentId: "cmt-1",
		AuthorId:  "user-1",
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

// DELETE 0 ligne et le commentaire existe → CommentForbidden.
func TestDeleteComment_Forbidden(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`DELETE FROM quote_line_comments`).
		WithArgs("cmt-1", "autre-user").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("cmt-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	resp, err := srv.DeleteComment(context.Background(), &quoteGrpc.DeleteCommentRequest{
		CommentId: "cmt-1",
		AuthorId:  "autre-user",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for wrong author")
	}
	if resp.Code != actions.CodeCommentForbidden {
		t.Fatalf("expected CodeCommentForbidden, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// DELETE 0 ligne et le commentaire n'existe pas → NotFound.
func TestDeleteComment_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`DELETE FROM quote_line_comments`).
		WithArgs("cmt-inexistant", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("cmt-inexistant").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	resp, err := srv.DeleteComment(context.Background(), &quoteGrpc.DeleteCommentRequest{
		CommentId: "cmt-inexistant",
		AuthorId:  "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for non-existent comment")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteComment_MissingFields(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.DeleteComment(context.Background(), &quoteGrpc.DeleteCommentRequest{
		CommentId: "cmt-1",
		// AuthorId manquant
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for missing author_id")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}
