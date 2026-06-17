package line

import (
	"context"
	"database/sql"
	"strconv"

	"project-devis-quote/actions/codes"
	"project-devis-quote/actions/quote"
	"project-devis-quote/actions/sqlutil"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *quoteGrpc.UpdateQuoteLineRequest) (*quoteGrpc.UpdateQuoteLineResponse, error) {
	var fieldErrors []*quoteGrpc.ValidationError

	if req.LineId == "" {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "line_id", Message: "Champ requis."})
	}
	if req.UserId == "" {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "user_id", Message: "Champ requis."})
	}
	if req.Type == "" {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "type", Message: "Champ requis."})
	} else if req.Type != TypeSimple && req.Type != TypeMultiple {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "type", Message: "Type invalide."})
	}
	if req.Quantity == "" {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "quantity", Message: "Champ requis."})
	} else if _, err := strconv.ParseFloat(req.Quantity, 64); err != nil {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "quantity", Message: "Doit être un nombre valide."})
	}
	if req.UnitPrice < 0 {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "unit_price", Message: "Doit être positif ou nul."})
	}

	if len(fieldErrors) > 0 {
		return &quoteGrpc.UpdateQuoteLineResponse{Success: false, Code: codes.InvalidInput, ValidationErrors: fieldErrors}, nil
	}

	cleanData, err := ValidateData(req.Type, req.Data)
	if err != nil {
		return &quoteGrpc.UpdateQuoteLineResponse{Success: false, Code: codes.InvalidLineData}, nil
	}

	if code, ok := quote.LineParentEditable(ctx, db, req.LineId, req.UserId); !ok {
		return &quoteGrpc.UpdateQuoteLineResponse{Success: false, Code: code}, nil
	}

	res, err := db.ExecContext(ctx,
		`UPDATE quote_lines
		 SET type=$1, name=$2, quantity=$3::DECIMAL, unit=$4, unit_price=$5, data=$6::jsonb,
		     position=$7, tax_id=$8, fee_id=$9, updated_at=NOW()
		 WHERE line_id=$10`,
		req.Type, req.Name, req.Quantity, sqlutil.NullableStr(req.Unit),
		req.UnitPrice, cleanData, req.Position,
		sqlutil.NullableInt32(req.TaxId),
		sqlutil.NullableStr(FeeIDFromData(cleanData)),
		req.LineId,
	)
	if err != nil {
		return &quoteGrpc.UpdateQuoteLineResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &quoteGrpc.UpdateQuoteLineResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &quoteGrpc.UpdateQuoteLineResponse{Success: true, Code: codes.Success}, nil
}
