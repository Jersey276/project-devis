package quote

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Get(ctx context.Context, db *sql.DB, req *quoteGrpc.GetQuoteRequest) (*quoteGrpc.GetQuoteResponse, error) {
	if req.QuoteId == "" || req.UserId == "" {
		return &quoteGrpc.GetQuoteResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var (
		quoteID       string
		userID        string
		name          string
		archivedAt    sql.NullTime
		state         string
		clientID      string
		addressID     int32
		userAddressID int32
		issuedAt      sql.NullTime
		validUntil    sql.NullTime
		paymentTerms  sql.NullString
	)
	err := db.QueryRowContext(ctx,
		`SELECT quote_id, user_id, name, archived_at, state, client_id, address_id, COALESCE(user_address_id, 0),
		        issued_at, valid_until, payment_terms
		 FROM quotes WHERE quote_id=$1 AND user_id=$2`,
		req.QuoteId, req.UserId,
	).Scan(&quoteID, &userID, &name, &archivedAt, &state, &clientID, &addressID, &userAddressID,
		&issuedAt, &validUntil, &paymentTerms)
	if err == sql.ErrNoRows {
		return &quoteGrpc.GetQuoteResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &quoteGrpc.GetQuoteResponse{Success: false, Code: codes.InternalError}, err
	}

	rows, err := db.QueryContext(ctx,
		`SELECT line_id, quote_id, type, name, quantity::text, COALESCE(unit, ''), unit_price, data::text, position, COALESCE(tax_id, 0)
		 FROM quote_lines WHERE quote_id=$1 ORDER BY position ASC, id ASC`,
		quoteID,
	)
	if err != nil {
		return &quoteGrpc.GetQuoteResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	var lines []*quoteGrpc.QuoteLine
	for rows.Next() {
		l := &quoteGrpc.QuoteLine{}
		if err := rows.Scan(&l.LineId, &l.QuoteId, &l.Type, &l.Name, &l.Quantity, &l.Unit, &l.UnitPrice, &l.Data, &l.Position, &l.TaxId); err != nil {
			return &quoteGrpc.GetQuoteResponse{Success: false, Code: codes.InternalError}, err
		}
		lines = append(lines, l)
	}
	if err := rows.Err(); err != nil {
		return &quoteGrpc.GetQuoteResponse{Success: false, Code: codes.InternalError}, err
	}

	issuedAtStr := ""
	if issuedAt.Valid {
		issuedAtStr = issuedAt.Time.UTC().Format("2006-01-02T15:04:05Z")
	}
	validUntilStr := ""
	if validUntil.Valid {
		validUntilStr = validUntil.Time.Format("2006-01-02")
	}

	return &quoteGrpc.GetQuoteResponse{
		Success: true,
		Code:    codes.Success,
		Quote: &quoteGrpc.Quote{
			QuoteId:       quoteID,
			UserId:        userID,
			Name:          name,
			Archived:      archivedAt.Valid,
			State:         StateFromString(state),
			ClientId:      clientID,
			AddressId:     addressID,
			UserAddressId: userAddressID,
			IssuedAt:      issuedAtStr,
			ValidUntil:    validUntilStr,
			PaymentTerms:  paymentTerms.String,
		},
		Lines: lines,
	}, nil
}
