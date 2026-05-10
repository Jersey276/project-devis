package sqlutil

import (
	"errors"
	"strconv"

	usersGrpc "project-devis-users/services/grpc"
)

// DB-side string equivalents of the OwnerType enum. These match the CHECK
// constraint in migration 000007 and the values stored in addresses.owner_type.
const (
	OwnerTypeUser   = "user"
	OwnerTypeClient = "client"
)

var ErrInvalidOwnerType = errors.New("invalid owner_type")

// OwnerTypeToDBString converts the proto enum to its DB string equivalent.
// OWNER_TYPE_UNSPECIFIED (and any unknown variant) returns ErrInvalidOwnerType.
func OwnerTypeToDBString(t usersGrpc.OwnerType) (string, error) {
	switch t {
	case usersGrpc.OwnerType_OWNER_TYPE_USER:
		return OwnerTypeUser, nil
	case usersGrpc.OwnerType_OWNER_TYPE_CLIENT:
		return OwnerTypeClient, nil
	default:
		return "", ErrInvalidOwnerType
	}
}

// OwnerTypeFromDBString is the inverse of OwnerTypeToDBString. Used when
// scanning a row from the addresses table back into a proto Address.
func OwnerTypeFromDBString(s string) (usersGrpc.OwnerType, error) {
	switch s {
	case OwnerTypeUser:
		return usersGrpc.OwnerType_OWNER_TYPE_USER, nil
	case OwnerTypeClient:
		return usersGrpc.OwnerType_OWNER_TYPE_CLIENT, nil
	default:
		return usersGrpc.OwnerType_OWNER_TYPE_UNSPECIFIED, ErrInvalidOwnerType
	}
}

// AddressAuthPredicate returns the SQL ownership-check fragment for the
// addresses table, with the auth-user placeholder bound to $authIdx. The
// caller embeds it inside a WHERE clause and passes the auth user id at the
// matching arg index. Centralised so the user-vs-client + archived-client
// rules stay in lock-step across get/list/create/update/archive.
//
//	WHERE id=$1 AND archived_at IS NULL
//	  AND ` + sqlutil.AddressAuthPredicate(2)
func AddressAuthPredicate(authIdx int) string {
	p := "$" + strconv.Itoa(authIdx)
	return `(
		(owner_type = 'user'   AND owner_id = ` + p + `)
		OR
		(owner_type = 'client' AND owner_id IN (
			SELECT client_id FROM clients
			WHERE user_id = ` + p + ` AND archived_at IS NULL
		))
	)`
}
