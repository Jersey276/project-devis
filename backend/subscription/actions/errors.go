package actions

import (
	"database/sql"
	"errors"

	"project-devis-subscription/actions/codes"
)

const (
	CodeSuccess       = codes.Success
	CodeNotFound      = codes.NotFound
	CodeAlreadyExists = codes.AlreadyExists
	CodeInvalidInput  = codes.InvalidInput
	CodeInternalError = codes.InternalError
)

// queryErrCode maps a DB query error to a response code.
// Returns (code, true) when err is non-nil, (0, false) when err is nil.
func queryErrCode(err error) (int32, bool) {
	if err == nil {
		return 0, false
	}
	if errors.Is(err, sql.ErrNoRows) {
		return CodeNotFound, true
	}
	return CodeInternalError, true
}
