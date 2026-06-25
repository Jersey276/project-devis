package comment

import (
	"context"
	"database/sql"
)

func commentExists(ctx context.Context, db *sql.DB, commentId string) bool {
	var exists bool
	_ = db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM quote_line_comments WHERE comment_id = $1)`,
		commentId,
	).Scan(&exists)
	return exists
}
