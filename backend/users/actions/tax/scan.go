package tax

import (
	"database/sql"

	usersGrpc "project-devis-users/services/grpc"
)

// Columns is the canonical SELECT list for the taxes table. Adding a new
// tax field means updating this constant and ScanRow/ScanRows together —
// no other call site should hand-roll the column list.
const Columns = `id, name, rate::TEXT, country_group_id, is_default,
	COALESCE(original_tax_id, 0), version,
	COALESCE(TO_CHAR(superseded_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), ''),
	COALESCE(superseded_by, 0)`

// row is the minimal Scan-able interface implemented by *sql.Row and *sql.Rows.
type row interface {
	Scan(dest ...interface{}) error
}

func ScanRow(r row) (*usersGrpc.Tax, error) {
	var t usersGrpc.Tax
	if err := r.Scan(
		&t.Id, &t.Name, &t.Rate, &t.CountryGroupId, &t.IsDefault,
		&t.OriginalTaxId, &t.Version, &t.SupersededAt, &t.SupersededBy,
	); err != nil {
		return nil, err
	}
	return &t, nil
}

func ScanRows(rows *sql.Rows) ([]*usersGrpc.Tax, error) {
	var taxes []*usersGrpc.Tax
	for rows.Next() {
		t, err := ScanRow(rows)
		if err != nil {
			return nil, err
		}
		taxes = append(taxes, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return taxes, nil
}
