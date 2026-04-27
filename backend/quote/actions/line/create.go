package line

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/google/uuid"
	"project-devis-quote/actions/codes"
	"project-devis-quote/actions/sqlutil"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *quoteGrpc.CreateQuoteLineRequest) (*quoteGrpc.CreateQuoteLineResponse, error) {
	if req.QuoteId == "" || req.UserId == "" || req.Type == "" || req.Name == "" || req.Quantity == "" {
		return &quoteGrpc.CreateQuoteLineResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	if req.UnitPrice < 0 {
		return &quoteGrpc.CreateQuoteLineResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	if _, err := strconv.ParseFloat(req.Quantity, 64); err != nil {
		return &quoteGrpc.CreateQuoteLineResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	if req.Type != TypeSimple && req.Type != TypeMultiple {
		return &quoteGrpc.CreateQuoteLineResponse{Success: false, Code: codes.InvalidLineType}, nil
	}
	cleanData, err := ValidateData(req.Type, req.Data)
	if err != nil {
		return &quoteGrpc.CreateQuoteLineResponse{Success: false, Code: codes.InvalidLineData}, nil
	}

	// Verify the parent quote belongs to the user and is not archived.
	var ownerOK bool
	err = db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM quotes WHERE quote_id=$1 AND user_id=$2 AND archived_at IS NULL)`,
		req.QuoteId, req.UserId,
	).Scan(&ownerOK)
	if err != nil {
		return &quoteGrpc.CreateQuoteLineResponse{Success: false, Code: codes.InternalError}, err
	}
	if !ownerOK {
		return &quoteGrpc.CreateQuoteLineResponse{Success: false, Code: codes.NotFound}, nil
	}

	lineID := uuid.New().String()
	_, err = db.ExecContext(ctx,
		`INSERT INTO quote_lines (line_id, quote_id, type, name, quantity, unit, unit_price, data, position)
		 VALUES ($1, $2, $3, $4, $5::DECIMAL, $6, $7, $8::jsonb, $9)`,
		lineID, req.QuoteId, req.Type, req.Name, req.Quantity,
		sqlutil.NullableStr(req.Unit), req.UnitPrice, cleanData, req.Position,
	)
	if err != nil {
		return &quoteGrpc.CreateQuoteLineResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.CreateQuoteLineResponse{Success: true, Code: codes.Success, LineId: lineID}, nil
}
