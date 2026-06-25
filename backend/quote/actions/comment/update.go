package comment

import (
	"context"
	"database/sql"
	"errors"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *quoteGrpc.UpdateCommentRequest) (*quoteGrpc.UpdateCommentResponse, error) {
	if req.CommentId == "" || req.AuthorId == "" || req.Body == "" {
		return &quoteGrpc.UpdateCommentResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var c quoteGrpc.QuoteLineComment
	err := db.QueryRowContext(ctx,
		`UPDATE quote_line_comments
		 SET body = $1, updated_at = NOW()
		 WHERE comment_id = $2 AND author_id = $3
		 RETURNING comment_id, line_id, quote_id, author_id, author_name, body,
		           created_at::TEXT, updated_at::TEXT`,
		req.Body, req.CommentId, req.AuthorId,
	).Scan(&c.CommentId, &c.LineId, &c.QuoteId, &c.AuthorId, &c.AuthorName, &c.Body, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		// Either not found or not owned by this author.
		var exists bool
		_ = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM quote_line_comments WHERE comment_id = $1)`, req.CommentId).Scan(&exists)
		if exists {
			return &quoteGrpc.UpdateCommentResponse{Success: false, Code: codes.CommentForbidden}, nil
		}
		return &quoteGrpc.UpdateCommentResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &quoteGrpc.UpdateCommentResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.UpdateCommentResponse{Success: true, Code: codes.Success, Comment: &c}, nil
}
