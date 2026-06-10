package tax

import (
	"context"
	"database/sql"
	"strconv"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

type currentRow struct {
	id          int32
	name        string
	rate        string
	groupID     int32
	originalID  sql.NullInt32
	supersededA sql.NullTime
}

// Update mutates a tax. If name or rate changed, the row is preserved
// (superseded_at/by) and a new version is inserted in the same family —
// quote_lines that point at the old id keep their snapshot. If only
// is_default changed, the current row is updated in place.
func Update(ctx context.Context, db *sql.DB, req *usersGrpc.UpdateTaxRequest) (*usersGrpc.UpdateTaxResponse, error) {
	var fieldErrors []*usersGrpc.ValidationError

	if req.TaxId == 0 {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "tax_id", Message: "Champ requis."})
	}
	if req.Name == "" {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "name", Message: "Champ requis."})
	}
	if req.Rate == "" {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "rate", Message: "Champ requis."})
	} else if err := validateRate(req.Rate); err != nil {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "rate", Message: "Taux invalide (0–999.99)."})
	}

	if len(fieldErrors) > 0 {
		return &usersGrpc.UpdateTaxResponse{Success: false, Code: codes.InvalidInput, ValidationErrors: fieldErrors}, nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return &usersGrpc.UpdateTaxResponse{Success: false, Code: codes.InternalError}, err
	}
	defer tx.Rollback()

	cur, code, err := loadCurrent(ctx, tx, req.TaxId)
	if err != nil {
		return &usersGrpc.UpdateTaxResponse{Success: false, Code: code}, err
	}
	if code != codes.Success {
		return &usersGrpc.UpdateTaxResponse{Success: false, Code: code}, nil
	}

	var newID int32
	if req.Name == cur.name && ratesEqual(req.Rate, cur.rate) {
		newID, err = applyInPlace(ctx, tx, req, cur)
	} else {
		newID, err = createNewVersion(ctx, tx, req, cur)
	}
	if err != nil {
		return &usersGrpc.UpdateTaxResponse{Success: false, Code: codes.InternalError}, err
	}

	if err := tx.Commit(); err != nil {
		return &usersGrpc.UpdateTaxResponse{Success: false, Code: codes.InternalError}, err
	}
	return &usersGrpc.UpdateTaxResponse{Success: true, Code: codes.Success, TaxId: newID}, nil
}

func loadCurrent(ctx context.Context, tx *sql.Tx, id int32) (*currentRow, int32, error) {
	cur := &currentRow{id: id}
	err := tx.QueryRowContext(ctx,
		`SELECT name, rate::TEXT, country_group_id, original_tax_id, superseded_at
		   FROM taxes WHERE id=$1`,
		id,
	).Scan(&cur.name, &cur.rate, &cur.groupID, &cur.originalID, &cur.supersededA)
	if err == sql.ErrNoRows {
		return nil, codes.NotFound, nil
	}
	if err != nil {
		return nil, codes.InternalError, err
	}
	if cur.supersededA.Valid {
		// Superseded rows are immutable history.
		return nil, codes.InvalidInput, nil
	}
	return cur, codes.Success, nil
}

func applyInPlace(ctx context.Context, tx *sql.Tx, req *usersGrpc.UpdateTaxRequest, cur *currentRow) (int32, error) {
	if req.IsDefault {
		if err := clearDefaultInGroup(ctx, tx, cur.groupID, cur.id); err != nil {
			return 0, err
		}
	}
	if _, err := tx.ExecContext(ctx,
		"UPDATE taxes SET is_default=$1 WHERE id=$2",
		req.IsDefault, cur.id,
	); err != nil {
		return 0, err
	}
	return cur.id, nil
}

func createNewVersion(ctx context.Context, tx *sql.Tx, req *usersGrpc.UpdateTaxRequest, cur *currentRow) (int32, error) {
	familyID := cur.id
	if cur.originalID.Valid {
		familyID = cur.originalID.Int32
	}

	if req.IsDefault {
		if err := clearDefaultInGroup(ctx, tx, cur.groupID, 0); err != nil {
			return 0, err
		}
	}

	var nextVersion int32
	if err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(version), 0) + 1 FROM taxes
		  WHERE id=$1 OR original_tax_id=$1`,
		familyID,
	).Scan(&nextVersion); err != nil {
		return 0, err
	}

	var newID int32
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO taxes (name, rate, country_group_id, is_default, original_tax_id, version)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		req.Name, req.Rate, cur.groupID, req.IsDefault, familyID, nextVersion,
	).Scan(&newID); err != nil {
		return 0, err
	}

	if _, err := tx.ExecContext(ctx,
		"UPDATE taxes SET superseded_at=NOW(), superseded_by=$1, is_default=FALSE WHERE id=$2",
		newID, cur.id,
	); err != nil {
		return 0, err
	}
	return newID, nil
}

// ratesEqual normalises decimal strings so "20" and "20.00" compare equal.
// On parse failure we conservatively treat them as different — better to
// create a redundant version than to silently skip a real update.
func ratesEqual(a, b string) bool {
	fa, errA := strconv.ParseFloat(a, 64)
	fb, errB := strconv.ParseFloat(b, 64)
	if errA != nil || errB != nil {
		return false
	}
	return fa == fb
}
