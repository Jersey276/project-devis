package quote

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func List(ctx context.Context, db *sql.DB, req *quoteGrpc.ListQuotesRequest) (*quoteGrpc.ListQuotesResponse, error) {
	if req.UserId == "" {
		return &quoteGrpc.ListQuotesResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	where, args := buildQuoteFilters(req.UserId, req.IncludeArchived, req.Filters)

	var total int64
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM quotes"+where, args...).Scan(&total); err != nil {
		return &quoteGrpc.ListQuotesResponse{Success: false, Code: codes.InternalError}, err
	}

	orderBy := buildQuoteOrderBy(req.SortBy, req.SortDirection)

	args = append(args, pageSize, offset)
	n := len(args)
	query := fmt.Sprintf(
		`SELECT quote_id, user_id, name, archived_at, state, client_id, address_id, COALESCE(user_address_id, 0), valid_until
		 FROM quotes%s ORDER BY %s LIMIT $%d OFFSET $%d`,
		where, orderBy, n-1, n,
	)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return &quoteGrpc.ListQuotesResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	var quotes []*quoteGrpc.Quote
	for rows.Next() {
		var (
			quoteID       string
			userID        string
			name          string
			archivedAt    sql.NullTime
			state         string
			clientID      string
			addressID     int32
			userAddressID int32
			validUntil    sql.NullTime
		)
		if err := rows.Scan(&quoteID, &userID, &name, &archivedAt, &state, &clientID, &addressID, &userAddressID, &validUntil); err != nil {
			return &quoteGrpc.ListQuotesResponse{Success: false, Code: codes.InternalError}, err
		}
		validUntilStr := ""
		if validUntil.Valid {
			validUntilStr = validUntil.Time.Format("2006-01-02")
		}
		quotes = append(quotes, &quoteGrpc.Quote{
			QuoteId:       quoteID,
			UserId:        userID,
			Name:          name,
			Archived:      archivedAt.Valid,
			State:         StateFromString(state),
			ClientId:      clientID,
			AddressId:     addressID,
			UserAddressId: userAddressID,
			ValidUntil:    validUntilStr,
		})
	}
	if err := rows.Err(); err != nil {
		return &quoteGrpc.ListQuotesResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.ListQuotesResponse{Success: true, Code: codes.Success, Quotes: quotes, Total: total}, nil
}

var allowedQuoteSortColumns = map[string]string{
	"id":          "quote_id",
	"projectName": "name",
	"status":      "state",
	"totalTtc":    "created_at", // pas de colonne total en DB, fallback
	"created_at":  "created_at",
}

func buildQuoteOrderBy(sortBy, sortDirection string) string {
	col, ok := allowedQuoteSortColumns[sortBy]
	if !ok {
		col = "created_at"
	}
	if strings.ToUpper(sortDirection) == "ASC" {
		return col + " ASC"
	}
	return col + " DESC"
}

func buildQuoteFilters(userID string, includeArchived bool, f *quoteGrpc.QuoteFilters) (string, []interface{}) {
	args := []interface{}{userID}
	clauses := []string{"user_id = $1"}

	if !includeArchived {
		clauses = append(clauses, "archived_at IS NULL")
	}

	if f != nil {
		if f.Search != "" {
			args = append(args, "%"+f.Search+"%")
			clauses = append(clauses, fmt.Sprintf("name ILIKE $%d", len(args)))
		}
		if len(f.States) > 0 {
			placeholders := make([]string, len(f.States))
			for i, s := range f.States {
				args = append(args, s)
				placeholders[i] = fmt.Sprintf("$%d", len(args))
			}
			clauses = append(clauses, "state IN ("+strings.Join(placeholders, ",")+")")
		}
		if f.ClientId != "" {
			args = append(args, f.ClientId)
			clauses = append(clauses, fmt.Sprintf("client_id = $%d", len(args)))
		}
		if len(f.QuoteIds) > 0 {
			placeholders := make([]string, len(f.QuoteIds))
			for i, id := range f.QuoteIds {
				args = append(args, id)
				placeholders[i] = fmt.Sprintf("$%d", len(args))
			}
			clauses = append(clauses, "quote_id IN ("+strings.Join(placeholders, ",")+")")
		}
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}
