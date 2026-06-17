package fee

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"project-devis-quote/actions/codes"
	"project-devis-quote/actions/sqlutil"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *quoteGrpc.CreateFeeRequest) (*quoteGrpc.CreateFeeResponse, error) {
	fieldErrors := validateInput(req.Category, req.Name)
	if req.UserId == "" {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "user_id", Message: "Champ requis."})
	}
	if req.UnitPrice < 0 {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "unit_price", Message: "Doit être positif ou nul."})
	}
	if len(fieldErrors) > 0 {
		return &quoteGrpc.CreateFeeResponse{Success: false, Code: codes.InvalidInput, ValidationErrors: fieldErrors}, nil
	}

	feeID := uuid.New().String()
	_, err := db.ExecContext(ctx,
		`INSERT INTO fees (fee_id, user_id, category, name, unit, unit_price, tax_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		feeID, req.UserId, req.Category, req.Name,
		sqlutil.NullableStr(req.Unit), req.UnitPrice, sqlutil.NullableInt32(req.TaxId),
	)
	if err != nil {
		return &quoteGrpc.CreateFeeResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.CreateFeeResponse{Success: true, Code: codes.Success, FeeId: feeID}, nil
}
