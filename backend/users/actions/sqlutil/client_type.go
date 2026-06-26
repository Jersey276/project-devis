package sqlutil

import (
	usersGrpc "project-devis-users/services/grpc"
)

// DB-side string equivalents of the ClientType enum. These match the CHECK
// constraint in migration 000012 and the values stored in clients.client_type.
const (
	ClientTypeIndividual = "individual"
	ClientTypeBusiness   = "business"
)

// ClientTypeToDBString converts the proto enum to its DB string equivalent.
// CLIENT_TYPE_UNSPECIFIED (and any unset value) defaults to "individual" (B2C)
// so callers that don't yet send the field — the current front-end — produce a
// valid row.
func ClientTypeToDBString(t usersGrpc.ClientType) string {
	if t == usersGrpc.ClientType_CLIENT_TYPE_BUSINESS {
		return ClientTypeBusiness
	}
	return ClientTypeIndividual
}

// ClientTypeFromDBString is the inverse of ClientTypeToDBString. Used when
// scanning a row from the clients table back into a proto Client. An empty or
// NULL-sourced string maps to CLIENT_TYPE_UNSPECIFIED so legacy rows that were
// not yet backfilled don't fabricate a type. Values are guarded by the DB CHECK
// constraint, so an unrecognised string falls back to UNSPECIFIED.
func ClientTypeFromDBString(s string) usersGrpc.ClientType {
	switch s {
	case ClientTypeIndividual:
		return usersGrpc.ClientType_CLIENT_TYPE_INDIVIDUAL
	case ClientTypeBusiness:
		return usersGrpc.ClientType_CLIENT_TYPE_BUSINESS
	default:
		return usersGrpc.ClientType_CLIENT_TYPE_UNSPECIFIED
	}
}
