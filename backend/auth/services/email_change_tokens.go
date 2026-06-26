package services

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

const EmailChangeTokenTTL = 24 * time.Hour

var (
	ErrEmailChangeTokenNotFound = errors.New("email change token not found")
	ErrEmailChangeTokenExpired  = errors.New("email change token expired")
	ErrEmailChangeTokenUsed     = errors.New("email change token already used")
)

func GenerateEmailChangeToken(ctx context.Context, db *sql.DB, userID, newEmail string) (string, error) {
	rawToken, err := generateRawToken()
	if err != nil {
		return "", err
	}
	tokenHash := hashToken(rawToken)
	expiresAt := time.Now().Add(EmailChangeTokenTTL)

	// Invalidate any existing pending token for this user before inserting.
	_, _ = db.ExecContext(ctx,
		"DELETE FROM email_change_tokens WHERE user_id = $1 AND used_at IS NULL",
		userID,
	)

	_, err = db.ExecContext(ctx,
		"INSERT INTO email_change_tokens (user_id, new_email, token_hash, expires_at) VALUES ($1, $2, $3, $4)",
		userID, newEmail, tokenHash, expiresAt,
	)
	if err != nil {
		return "", err
	}
	return rawToken, nil
}

// ValidateEmailChangeToken checks the token and returns (userID, newEmail).
func ValidateEmailChangeToken(ctx context.Context, db *sql.DB, rawToken string) (string, string, error) {
	tokenHash := hashToken(rawToken)

	var userID, newEmail string
	var expiresAt time.Time
	var usedAt sql.NullTime
	err := db.QueryRowContext(ctx,
		"SELECT user_id, new_email, expires_at, used_at FROM email_change_tokens WHERE token_hash = $1",
		tokenHash,
	).Scan(&userID, &newEmail, &expiresAt, &usedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", ErrEmailChangeTokenNotFound
		}
		return "", "", err
	}

	if usedAt.Valid {
		return "", "", ErrEmailChangeTokenUsed
	}
	if time.Now().After(expiresAt) {
		return "", "", ErrEmailChangeTokenExpired
	}
	return userID, newEmail, nil
}

func ConsumeEmailChangeToken(ctx context.Context, db *sql.DB, rawToken string) error {
	tokenHash := hashToken(rawToken)
	result, err := db.ExecContext(ctx,
		"UPDATE email_change_tokens SET used_at = NOW() WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()",
		tokenHash,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrEmailChangeTokenNotFound
	}
	return nil
}
