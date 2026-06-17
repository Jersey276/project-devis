package fee

import (
	"context"
	"database/sql"
	"encoding/json"

	"project-devis-quote/actions/quote"
	"project-devis-quote/actions/sqlutil"
)

// editableStates are the quote states whose lines may still be rewritten when a
// referenced fee changes — the canonical set lives in quote.EditableStates
// (draft/sent, non-archived). Validated and drop quotes are immutable.
var editableStates = quote.EditableStates()

// feeSnapshot is the set of fields copied from a fee onto every line that
// references it.
type feeSnapshot struct {
	Name      string
	Unit      string
	UnitPrice int64
	TaxID     int32
}

// propagate rewrites the snapshot of every quote line referencing feeID, but
// only for lines belonging to editable (draft/sent, non-archived) quotes owned
// by userID. It covers both top-level fee lines (kind="fee", tracked via the
// fee_id column) and fee sublines nested inside detailed (multiple) lines.
//
// Propagation is best-effort relative to the fee update: the caller logs but
// does not fail the update if this returns an error, so a fee edit always
// succeeds even if a quote could not be refreshed.
func propagate(ctx context.Context, db *sql.DB, userID, feeID string, snap feeSnapshot) error {
	if err := propagateLines(ctx, db, userID, feeID, snap); err != nil {
		return err
	}
	return propagateSublines(ctx, db, userID, feeID, snap)
}

// propagateLines updates top-level fee lines via the indexed fee_id column.
func propagateLines(ctx context.Context, db *sql.DB, userID, feeID string, snap feeSnapshot) error {
	_, err := db.ExecContext(ctx,
		`UPDATE quote_lines l
		 SET name=$1, unit=$2, unit_price=$3, tax_id=$4, updated_at=NOW()
		 FROM quotes q
		 WHERE l.quote_id = q.quote_id
		   AND l.fee_id = $5
		   AND q.user_id = $6
		   AND q.archived_at IS NULL
		   AND q.state = ANY($7)`,
		snap.Name, sqlutil.NullableStr(snap.Unit), snap.UnitPrice,
		sqlutil.NullableInt32(snap.TaxID), feeID, userID, sqlutil.StringArray(editableStates),
	)
	return err
}

type sublinePayload struct {
	Name      string  `json:"name"`
	Quantity  string  `json:"quantity"`
	Unit      *string `json:"unit,omitempty"`
	UnitPrice int64   `json:"unit_price"`
	Option    *bool   `json:"option,omitempty"`
	FeeID     string  `json:"fee_id,omitempty"`
}

type multipleData struct {
	Kind        string           `json:"kind,omitempty"`
	Description string           `json:"description,omitempty"`
	Sublines    []sublinePayload `json:"sublines,omitempty"`
}

// propagateSublines refreshes fee sublines nested inside detailed lines. Because
// the fee reference lives inside the JSONB sublines array, each candidate line
// is read, patched in Go, and written back. Only lines of editable quotes whose
// data JSON contains the fee_id are considered.
func propagateSublines(ctx context.Context, db *sql.DB, userID, feeID string, snap feeSnapshot) error {
	rows, err := db.QueryContext(ctx,
		`SELECT l.line_id, l.data::text
		 FROM quote_lines l
		 JOIN quotes q ON q.quote_id = l.quote_id
		 WHERE l.type = 'multiple'
		   AND q.user_id = $1
		   AND q.archived_at IS NULL
		   AND q.state = ANY($2)
		   AND l.data @> $3::jsonb`,
		userID, sqlutil.StringArray(editableStates),
		sublineMatch(feeID),
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	var lineIDs, datas []string
	for rows.Next() {
		var lineID, raw string
		if err := rows.Scan(&lineID, &raw); err != nil {
			return err
		}
		var data multipleData
		if err := json.Unmarshal([]byte(raw), &data); err != nil {
			continue
		}
		changed := false
		for i := range data.Sublines {
			if data.Sublines[i].FeeID != feeID {
				continue
			}
			data.Sublines[i].Name = snap.Name
			data.Sublines[i].UnitPrice = snap.UnitPrice
			if snap.Unit == "" {
				data.Sublines[i].Unit = nil
			} else {
				u := snap.Unit
				data.Sublines[i].Unit = &u
			}
			changed = true
		}
		if !changed {
			continue
		}
		clean, err := json.Marshal(data)
		if err != nil {
			continue
		}
		lineIDs = append(lineIDs, lineID)
		datas = append(datas, string(clean))
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(lineIDs) == 0 {
		return nil
	}

	// Single round-trip: zip the patched (line_id, data) pairs into a derived
	// table and join on line_id, instead of one UPDATE per row.
	_, err = db.ExecContext(ctx,
		`UPDATE quote_lines l
		 SET data = p.data::jsonb, updated_at = NOW()
		 FROM (SELECT unnest($1::text[]) AS line_id, unnest($2::text[]) AS data) p
		 WHERE l.line_id = p.line_id`,
		sqlutil.StringArray(lineIDs), sqlutil.StringArray(datas),
	)
	return err
}

// sublineMatch builds the jsonb containment operand that pre-filters detailed
// lines to those whose sublines reference feeID, so we only read candidate rows.
func sublineMatch(feeID string) string {
	b, _ := json.Marshal(map[string]any{
		"sublines": []map[string]string{{"fee_id": feeID}},
	})
	return string(b)
}
