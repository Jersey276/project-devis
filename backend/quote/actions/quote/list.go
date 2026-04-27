package quote

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func List(ctx context.Context, db *sql.DB, req *quoteGrpc.ListQuotesRequest) (*quoteGrpc.ListQuotesResponse, error) {
	if req.UserId == "" {
		return &quoteGrpc.ListQuotesResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	query := `SELECT quote_id, user_id, name, archived_at, state FROM quotes WHERE user_id=$1`
	if !req.IncludeArchived {
		query += ` AND archived_at IS NULL`
	}
	query += ` ORDER BY created_at DESC`

	rows, err := db.QueryContext(ctx, query, req.UserId)
	if err != nil {
		return &quoteGrpc.ListQuotesResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	var quotes []*quoteGrpc.Quote
	for rows.Next() {
		var (
			quoteID    string
			userID     string
			name       string
			archivedAt sql.NullTime
			state      string
		)
		if err := rows.Scan(&quoteID, &userID, &name, &archivedAt, &state); err != nil {
			return &quoteGrpc.ListQuotesResponse{Success: false, Code: codes.InternalError}, err
		}
		quotes = append(quotes, &quoteGrpc.Quote{
			QuoteId:  quoteID,
			UserId:   userID,
			Name:     name,
			Archived: archivedAt.Valid,
			State:    StateFromString(state),
		})
	}
	if err := rows.Err(); err != nil {
		return &quoteGrpc.ListQuotesResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.ListQuotesResponse{Success: true, Code: codes.Success, Quotes: quotes}, nil
}
