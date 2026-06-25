package comment

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *quoteGrpc.CreateCommentRequest) (*quoteGrpc.CreateCommentResponse, error) {
	if req.LineId == "" || req.QuoteId == "" || req.AuthorId == "" || req.Body == "" {
		return &quoteGrpc.CreateCommentResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	if req.AuthorName == "" {
		req.AuthorName = "Inconnu"
	}

	var c quoteGrpc.QuoteLineComment
	err := db.QueryRowContext(ctx,
		`INSERT INTO quote_line_comments (line_id, quote_id, author_id, author_name, body)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING comment_id, line_id, quote_id, author_id, author_name, body,
		           created_at::TEXT, updated_at::TEXT`,
		req.LineId, req.QuoteId, req.AuthorId, req.AuthorName, req.Body,
	).Scan(&c.CommentId, &c.LineId, &c.QuoteId, &c.AuthorId, &c.AuthorName, &c.Body, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return &quoteGrpc.CreateCommentResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.CreateCommentResponse{Success: true, Code: codes.Success, Comment: &c}, nil
}
