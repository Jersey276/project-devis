package quote

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

const (
	StateDraft       = "draft"
	StateNegociation = "negociation"
	StateValidated   = "validated"
	StateDrop        = "drop"
)

func EditableStates() []string {
	return []string{StateDraft, StateSent}
}

func StateFromString(s string) quoteGrpc.QuoteState {
	switch s {
	case StateDraft:
		return quoteGrpc.QuoteState_QUOTE_STATE_DRAFT
	case StateNegociation:
		return quoteGrpc.QuoteState_QUOTE_STATE_NEGOCIATION
	case StateValidated:
		return quoteGrpc.QuoteState_QUOTE_STATE_VALIDATED
	case StateDrop:
		return quoteGrpc.QuoteState_QUOTE_STATE_DROP
	default:
		return quoteGrpc.QuoteState_QUOTE_STATE_UNSPECIFIED
	}
}

// EditableForUser checks the parent quote belongs to the user and is in an
// editable state (not archived, not validated, not drop).
// Returns (code, ok). When ok is false, code is the response code to surface.
func EditableForUser(ctx context.Context, db *sql.DB, quoteID, userID string) (int32, bool) {
	var state string
	var archivedAt sql.NullTime
	err := db.QueryRowContext(ctx,
		`SELECT state, archived_at FROM quotes WHERE quote_id=$1 AND user_id=$2`,
		quoteID, userID,
	).Scan(&state, &archivedAt)
	if err == sql.ErrNoRows {
		return codes.NotFound, false
	}
	if err != nil {
		return codes.InternalError, false
	}
	return classifyEditable(state, archivedAt.Valid)
}

func classifyEditable(state string, archived bool) (int32, bool) {
	if archived {
		return codes.NotFound, false
	}
	if state == StateValidated || state == StateDrop {
		return codes.QuoteFinalized, false
	}
	return codes.Success, true
}

// LineParentEditable resolves the parent quote of a line and applies the same
// guard as EditableForUser in a single round-trip.
func LineParentEditable(ctx context.Context, db *sql.DB, lineID, userID string) (int32, bool) {
	var state string
	var archivedAt sql.NullTime
	err := db.QueryRowContext(ctx,
		`SELECT q.state, q.archived_at
		 FROM quote_lines l
		 JOIN quotes q ON q.quote_id = l.quote_id
		 WHERE l.line_id=$1 AND q.user_id=$2`,
		lineID, userID,
	).Scan(&state, &archivedAt)
	if err == sql.ErrNoRows {
		return codes.NotFound, false
	}
	if err != nil {
		return codes.InternalError, false
	}
	return classifyEditable(state, archivedAt.Valid)
}
