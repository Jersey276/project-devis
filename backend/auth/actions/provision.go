package actions

import (
	"context"
	"database/sql"

	userGrpc "project-devis-auth/services/user_auth"
)

// provisionUser creates a users-service user (admin if it is the very first
// account) and the matching auth row under an advisory lock, returning the new
// user_id. hashedPassword may be "" for OAuth-only accounts (auth.password is
// nullable). emailVerified sets the auth.email_verified column. On any failure
// after CreateUser it rolls the user back. Shared by Register (password set,
// email_verified=false) and OAuthLogin (password NULL, email_verified=true).
func (s *Server) provisionUser(ctx context.Context, normalizedEmail, hashedPassword string, emailVerified bool) (string, error) {
	// Pre-check: is this the very first registration? Used to set the admin role
	// in the users service at creation time. The transaction below re-verifies
	// under an advisory lock, so auth.role is always race-free.
	var preCount int64
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM auth").Scan(&preCount); err != nil {
		return "", err
	}
	isFirstUser := preCount == 0

	insertResp, err := s.userClient.CreateUser(ctx, &userGrpc.CreateUserRequest{
		Email:   normalizedEmail,
		IsAdmin: isFirstUser,
	})
	if err != nil {
		return "", err
	}
	if !insertResp.GetSuccess() {
		return "", &provisionError{code: insertResp.GetCode()}
	}

	userID := insertResp.GetUserId()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.rollbackUser(ctx, userID)
		return "", err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", int64(2026052901)); err != nil {
		s.rollbackUser(ctx, userID)
		return "", err
	}

	role := "free_user"
	var authCount int64
	if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM auth").Scan(&authCount); err != nil {
		s.rollbackUser(ctx, userID)
		return "", err
	}
	if authCount == 0 {
		role = "super_admin"
	}

	var password sql.NullString
	if hashedPassword != "" {
		password = sql.NullString{String: hashedPassword, Valid: true}
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO auth (user_id, email, password, role, account_status, subscription_tier, email_verified) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		userID,
		normalizedEmail,
		password,
		role,
		"active",
		"free",
		emailVerified,
	)
	if err != nil {
		s.rollbackUser(ctx, userID)
		return "", err
	}

	if err = tx.Commit(); err != nil {
		s.rollbackUser(ctx, userID)
		return "", err
	}

	return userID, nil
}

// provisionError carries a users-service response code from provisionUser so the
// caller can surface it. A zero code is normalized to CodeUserServiceError.
type provisionError struct {
	code int32
}

func (e *provisionError) Error() string {
	return "user provisioning failed"
}

// Code returns the mapped response code, defaulting to CodeUserServiceError.
func (e *provisionError) Code() int32 {
	if e.code == 0 {
		return CodeUserServiceError
	}
	return e.code
}
