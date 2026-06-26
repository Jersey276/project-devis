package fee

import (
	"database/sql"

	quoteGrpc "project-devis-quote/services/grpc"
)

type rowScanner interface {
	Scan(dest ...any) error
}

// scanFee maps a fees row into the proto message. Column order must match the
// SELECT used by Get and List.
func scanFee(rs rowScanner) (*quoteGrpc.Fee, error) {
	var (
		feeID      string
		userID     string
		category   string
		name       string
		unit       string
		unitPrice  int64
		taxID      int32
		archivedAt sql.NullTime
	)
	if err := rs.Scan(&feeID, &userID, &category, &name, &unit, &unitPrice, &taxID, &archivedAt); err != nil {
		return nil, err
	}
	return &quoteGrpc.Fee{
		FeeId:     feeID,
		UserId:    userID,
		Category:  category,
		Name:      name,
		Unit:      unit,
		UnitPrice: unitPrice,
		TaxId:     taxID,
		Archived:  archivedAt.Valid,
	}, nil
}

const selectColumns = `fee_id, user_id, category, name, COALESCE(unit, ''), unit_price, COALESCE(tax_id, 0), archived_at`
