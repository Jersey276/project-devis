package line

import (
	"context"
	"database/sql"
	"strconv"

	"project-devis-quote/actions/codes"
	"project-devis-quote/actions/sqlutil"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *quoteGrpc.UpdateQuoteLineRequest) (*quoteGrpc.UpdateQuoteLineResponse, error) {
	if req.LineId == "" || req.UserId == "" || req.Type == "" || req.Name == "" || req.Quantity == "" {
		return &quoteGrpc.UpdateQuoteLineResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	if req.UnitPrice < 0 {
		return &quoteGrpc.UpdateQuoteLineResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	if _, err := strconv.ParseFloat(req.Quantity, 64); err != nil {
		return &quoteGrpc.UpdateQuoteLineResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	if req.Type != TypeSimple && req.Type != TypeMultiple {
		return &quoteGrpc.UpdateQuoteLineResponse{Success: false, Code: codes.InvalidLineType}, nil
	}
	cleanData, err := ValidateData(req.Type, req.Data)
	if err != nil {
		return &quoteGrpc.UpdateQuoteLineResponse{Success: false, Code: codes.InvalidLineData}, nil
	}

	res, err := db.ExecContext(ctx,
		`UPDATE quote_lines l
		 SET type=$1, name=$2, quantity=$3::DECIMAL, unit=$4, unit_price=$5, data=$6::jsonb,
		     position=$7, updated_at=NOW()
		 FROM quotes q
		 WHERE l.quote_id = q.quote_id
		   AND l.line_id = $8
		   AND q.user_id = $9
		   AND q.archived_at IS NULL`,
		req.Type, req.Name, req.Quantity, sqlutil.NullableStr(req.Unit),
		req.UnitPrice, cleanData, req.Position,
		req.LineId, req.UserId,
	)
	if err != nil {
		return &quoteGrpc.UpdateQuoteLineResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &quoteGrpc.UpdateQuoteLineResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &quoteGrpc.UpdateQuoteLineResponse{Success: true, Code: codes.Success}, nil
}
