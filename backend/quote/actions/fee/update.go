package fee

import (
	"context"
	"database/sql"
	"log"

	"project-devis-quote/actions/codes"
	"project-devis-quote/actions/sqlutil"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *quoteGrpc.UpdateFeeRequest) (*quoteGrpc.UpdateFeeResponse, error) {
	fieldErrors := validateInput(req.Category, req.Name)
	if req.FeeId == "" {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "fee_id", Message: "Champ requis."})
	}
	if req.UserId == "" {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "user_id", Message: "Champ requis."})
	}
	if req.UnitPrice < 0 {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "unit_price", Message: "Doit être positif ou nul."})
	}
	if len(fieldErrors) > 0 {
		return &quoteGrpc.UpdateFeeResponse{Success: false, Code: codes.InvalidInput, ValidationErrors: fieldErrors}, nil
	}

	res, err := db.ExecContext(ctx,
		`UPDATE fees
		 SET category=$1, name=$2, unit=$3, unit_price=$4, tax_id=$5, updated_at=NOW()
		 WHERE fee_id=$6 AND user_id=$7 AND archived_at IS NULL`,
		req.Category, req.Name, sqlutil.NullableStr(req.Unit), req.UnitPrice,
		sqlutil.NullableInt32(req.TaxId), req.FeeId, req.UserId,
	)
	if err != nil {
		return &quoteGrpc.UpdateFeeResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &quoteGrpc.UpdateFeeResponse{Success: false, Code: codes.NotFound}, nil
	}

	// Propagate the new snapshot to every non-validated quote that references
	// this fee. Best-effort: a propagation failure must not fail the fee update.
	snap := feeSnapshot{Name: req.Name, Unit: req.Unit, UnitPrice: req.UnitPrice}
	if err := propagate(ctx, db, req.UserId, req.FeeId, snap); err != nil {
		log.Printf("fee.Update: propagation failed for fee %s: %v", req.FeeId, err)
	}

	return &quoteGrpc.UpdateFeeResponse{Success: true, Code: codes.Success}, nil
}
