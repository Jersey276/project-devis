package line

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func ListForUser(ctx context.Context, db *sql.DB, req *quoteGrpc.ListUserQuoteLinesRequest) (*quoteGrpc.ListUserQuoteLinesResponse, error) {
	if req.UserId == "" {
		return &quoteGrpc.ListUserQuoteLinesResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	query := `SELECT ql.line_id, ql.quote_id, ql.type, ql.name, ql.quantity::text, COALESCE(ql.unit, ''), ql.unit_price, ql.data::text, ql.position, COALESCE(ql.tax_id, 0)
	          FROM quote_lines ql
	          JOIN quotes q ON q.quote_id = ql.quote_id
	          WHERE q.user_id = $1`
	if !req.IncludeArchived {
		query += ` AND q.archived_at IS NULL`
	}
	query += ` ORDER BY ql.quote_id, ql.position ASC, ql.id ASC`

	rows, err := db.QueryContext(ctx, query, req.UserId)
	if err != nil {
		return &quoteGrpc.ListUserQuoteLinesResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	var lines []*quoteGrpc.QuoteLine
	for rows.Next() {
		l := &quoteGrpc.QuoteLine{}
		if err := rows.Scan(&l.LineId, &l.QuoteId, &l.Type, &l.Name, &l.Quantity, &l.Unit, &l.UnitPrice, &l.Data, &l.Position, &l.TaxId); err != nil {
			return &quoteGrpc.ListUserQuoteLinesResponse{Success: false, Code: codes.InternalError}, err
		}
		lines = append(lines, l)
	}
	if err := rows.Err(); err != nil {
		return &quoteGrpc.ListUserQuoteLinesResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.ListUserQuoteLinesResponse{Success: true, Code: codes.Success, Lines: lines}, nil
}
