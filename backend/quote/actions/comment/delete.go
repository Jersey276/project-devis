package comment

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Delete(ctx context.Context, db *sql.DB, req *quoteGrpc.DeleteCommentRequest) (*quoteGrpc.GenericResponse, error) {
	if req.CommentId == "" || req.AuthorId == "" {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	res, err := db.ExecContext(ctx,
		`DELETE FROM quote_line_comments WHERE comment_id = $1 AND author_id = $2`,
		req.CommentId, req.AuthorId,
	)
	if err != nil {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		if commentExists(ctx, db, req.CommentId) {
			return &quoteGrpc.GenericResponse{Success: false, Code: codes.CommentForbidden}, nil
		}
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &quoteGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
