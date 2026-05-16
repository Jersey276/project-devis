package tax

import (
	"context"
	"database/sql"
)

// execContext is the subset of *sql.DB / *sql.Tx we need for default
// management, so the helper can be called inside or outside a transaction.
type execContext interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// clearDefaultInGroup unsets is_default on every CURRENT tax in the group.
// excludeID=0 means "no exclusion". Superseded rows are left alone — they
// are immutable history and the unique partial index already restricts the
// "single default" invariant to current versions.
func clearDefaultInGroup(ctx context.Context, exec execContext, groupID, excludeID int32) error {
	if excludeID == 0 {
		_, err := exec.ExecContext(ctx,
			"UPDATE taxes SET is_default=FALSE WHERE country_group_id=$1 AND is_default=TRUE AND superseded_at IS NULL",
			groupID,
		)
		return err
	}
	_, err := exec.ExecContext(ctx,
		"UPDATE taxes SET is_default=FALSE WHERE country_group_id=$1 AND id<>$2 AND is_default=TRUE AND superseded_at IS NULL",
		groupID, excludeID,
	)
	return err
}
