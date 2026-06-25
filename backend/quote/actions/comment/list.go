package comment

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func List(ctx context.Context, db *sql.DB, req *quoteGrpc.ListCommentsRequest) (*quoteGrpc.ListCommentsResponse, error) {
	if req.LineId == "" {
		return &quoteGrpc.ListCommentsResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	rows, err := db.QueryContext(ctx,
		`SELECT comment_id, line_id, quote_id, author_id, author_name, body,
		        created_at::TEXT, updated_at::TEXT
		 FROM quote_line_comments
		 WHERE line_id = $1
		 ORDER BY created_at ASC`,
		req.LineId,
	)
	if err != nil {
		return &quoteGrpc.ListCommentsResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	comments := make([]*quoteGrpc.QuoteLineComment, 0)
	for rows.Next() {
		var c quoteGrpc.QuoteLineComment
		if err := rows.Scan(&c.CommentId, &c.LineId, &c.QuoteId, &c.AuthorId, &c.AuthorName, &c.Body, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return &quoteGrpc.ListCommentsResponse{Success: false, Code: codes.InternalError}, err
		}
		comments = append(comments, &c)
	}
	if err := rows.Err(); err != nil {
		return &quoteGrpc.ListCommentsResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.ListCommentsResponse{Success: true, Code: codes.Success, Comments: comments}, nil
}
