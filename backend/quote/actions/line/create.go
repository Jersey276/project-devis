package line

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/google/uuid"
	"project-devis-quote/actions/codes"
	"project-devis-quote/actions/quote"
	"project-devis-quote/actions/sqlutil"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *quoteGrpc.CreateQuoteLineRequest) (*quoteGrpc.CreateQuoteLineResponse, error) {
	var fieldErrors []*quoteGrpc.ValidationError

	if req.QuoteId == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("quote_id"))
	}
	if req.UserId == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("user_id"))
	}
	if req.Type == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("type"))
	} else if req.Type != sqlutil.TypeSimple && req.Type != sqlutil.TypeMultiple {
		fieldErrors = append(fieldErrors, sqlutil.Invalid("type", "Type invalide."))
	}
	if req.Quantity == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("quantity"))
	} else if _, err := strconv.ParseFloat(req.Quantity, 64); err != nil {
		fieldErrors = append(fieldErrors, sqlutil.Invalid("quantity", "Doit être un nombre valide."))
	}
	if req.UnitPrice < 0 {
		fieldErrors = append(fieldErrors, sqlutil.NonNegative("unit_price"))
	}

	if len(fieldErrors) > 0 {
		return &quoteGrpc.CreateQuoteLineResponse{Success: false, Code: codes.InvalidInput, ValidationErrors: fieldErrors}, nil
	}

	cleanData, err := ValidateData(req.Type, req.Data)
	if err != nil {
		return &quoteGrpc.CreateQuoteLineResponse{Success: false, Code: codes.InvalidLineData}, nil
	}

	if code, ok := quote.EditableForUser(ctx, db, req.QuoteId, req.UserId); !ok {
		return &quoteGrpc.CreateQuoteLineResponse{Success: false, Code: code}, nil
	}

	lineID := uuid.New().String()
	_, err = db.ExecContext(ctx,
		`INSERT INTO quote_lines (line_id, quote_id, type, name, quantity, unit, unit_price, data, position, tax_id, fee_id)
		 VALUES ($1, $2, $3, $4, $5::DECIMAL, $6, $7, $8::jsonb, $9, $10, $11)`,
		lineID, req.QuoteId, req.Type, req.Name, req.Quantity,
		sqlutil.NullableStr(req.Unit), req.UnitPrice, cleanData, req.Position,
		sqlutil.NullableInt32(req.TaxId), sqlutil.NullableStr(FeeIDFromData(cleanData)),
	)
	if err != nil {
		return &quoteGrpc.CreateQuoteLineResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.CreateQuoteLineResponse{Success: true, Code: codes.Success, LineId: lineID}, nil
}
