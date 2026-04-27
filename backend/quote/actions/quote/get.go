package quote

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Get(ctx context.Context, db *sql.DB, req *quoteGrpc.GetQuoteRequest) (*quoteGrpc.GetQuoteResponse, error) {
	if req.QuoteId == "" || req.UserId == "" {
		return &quoteGrpc.GetQuoteResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var (
		quoteID    string
		userID     string
		name       string
		archivedAt sql.NullTime
	)
	err := db.QueryRowContext(ctx,
		`SELECT quote_id, user_id, name, archived_at FROM quotes WHERE quote_id=$1 AND user_id=$2`,
		req.QuoteId, req.UserId,
	).Scan(&quoteID, &userID, &name, &archivedAt)
	if err == sql.ErrNoRows {
		return &quoteGrpc.GetQuoteResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &quoteGrpc.GetQuoteResponse{Success: false, Code: codes.InternalError}, err
	}

	rows, err := db.QueryContext(ctx,
		`SELECT line_id, quote_id, type, name, quantity::text, COALESCE(unit, ''), unit_price, data::text, position
		 FROM quote_lines WHERE quote_id=$1 ORDER BY position ASC, id ASC`,
		quoteID,
	)
	if err != nil {
		return &quoteGrpc.GetQuoteResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	var lines []*quoteGrpc.QuoteLine
	for rows.Next() {
		l := &quoteGrpc.QuoteLine{}
		if err := rows.Scan(&l.LineId, &l.QuoteId, &l.Type, &l.Name, &l.Quantity, &l.Unit, &l.UnitPrice, &l.Data, &l.Position); err != nil {
			return &quoteGrpc.GetQuoteResponse{Success: false, Code: codes.InternalError}, err
		}
		lines = append(lines, l)
	}
	if err := rows.Err(); err != nil {
		return &quoteGrpc.GetQuoteResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.GetQuoteResponse{
		Success: true,
		Code:    codes.Success,
		Quote: &quoteGrpc.Quote{
			QuoteId:  quoteID,
			UserId:   userID,
			Name:     name,
			Archived: archivedAt.Valid,
		},
		Lines: lines,
	}, nil
}
