package line

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Get(ctx context.Context, db *sql.DB, req *quoteGrpc.GetQuoteLineRequest) (*quoteGrpc.GetQuoteLineResponse, error) {
	if req.LineId == "" || req.UserId == "" {
		return &quoteGrpc.GetQuoteLineResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	l := &quoteGrpc.QuoteLine{}
	err := db.QueryRowContext(ctx,
		`SELECT l.line_id, l.quote_id, l.type, l.name, l.quantity::text, COALESCE(l.unit, ''), l.unit_price, l.data::text, l.position, COALESCE(l.tax_id, 0)
		 FROM quote_lines l
		 JOIN quotes q ON q.quote_id = l.quote_id
		 WHERE l.line_id=$1 AND q.user_id=$2`,
		req.LineId, req.UserId,
	).Scan(&l.LineId, &l.QuoteId, &l.Type, &l.Name, &l.Quantity, &l.Unit, &l.UnitPrice, &l.Data, &l.Position, &l.TaxId)
	if err == sql.ErrNoRows {
		return &quoteGrpc.GetQuoteLineResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &quoteGrpc.GetQuoteLineResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.GetQuoteLineResponse{Success: true, Code: codes.Success, Line: l}, nil
}
