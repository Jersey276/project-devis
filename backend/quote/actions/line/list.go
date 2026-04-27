package line

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func List(ctx context.Context, db *sql.DB, req *quoteGrpc.ListQuoteLinesRequest) (*quoteGrpc.ListQuoteLinesResponse, error) {
	if req.QuoteId == "" || req.UserId == "" {
		return &quoteGrpc.ListQuoteLinesResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var ownerOK bool
	if err := db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM quotes WHERE quote_id=$1 AND user_id=$2)`,
		req.QuoteId, req.UserId,
	).Scan(&ownerOK); err != nil {
		return &quoteGrpc.ListQuoteLinesResponse{Success: false, Code: codes.InternalError}, err
	}
	if !ownerOK {
		return &quoteGrpc.ListQuoteLinesResponse{Success: false, Code: codes.NotFound}, nil
	}

	rows, err := db.QueryContext(ctx,
		`SELECT line_id, quote_id, type, name, quantity::text, COALESCE(unit, ''), unit_price, data::text, position
		 FROM quote_lines WHERE quote_id=$1 ORDER BY position ASC, id ASC`,
		req.QuoteId,
	)
	if err != nil {
		return &quoteGrpc.ListQuoteLinesResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	var lines []*quoteGrpc.QuoteLine
	for rows.Next() {
		l := &quoteGrpc.QuoteLine{}
		if err := rows.Scan(&l.LineId, &l.QuoteId, &l.Type, &l.Name, &l.Quantity, &l.Unit, &l.UnitPrice, &l.Data, &l.Position); err != nil {
			return &quoteGrpc.ListQuoteLinesResponse{Success: false, Code: codes.InternalError}, err
		}
		lines = append(lines, l)
	}
	if err := rows.Err(); err != nil {
		return &quoteGrpc.ListQuoteLinesResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.ListQuoteLinesResponse{Success: true, Code: codes.Success, Lines: lines}, nil
}
