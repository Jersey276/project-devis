package services

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

const PasswordResetTokenTTL = 15 * time.Minute

var (
	ErrPasswordResetTokenNotFound = errors.New("password reset token not found")
	ErrPasswordResetTokenExpired  = errors.New("password reset token expired")
	ErrPasswordResetTokenUsed     = errors.New("password reset token already used")
)

func GeneratePasswordResetToken(ctx context.Context, db *sql.DB, userID string) (string, error) {
	rawToken, err := generateRawToken()
	if err != nil {
		return "", err
	}

	tokenHash := hashToken(rawToken)
	expiresAt := time.Now().Add(PasswordResetTokenTTL)

	_, err = db.ExecContext(
		ctx,
		"INSERT INTO password_reset_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)",
		userID,
		tokenHash,
		expiresAt,
	)
	if err != nil {
		return "", err
	}

	return rawToken, nil
}

func ValidatePasswordResetToken(ctx context.Context, db *sql.DB, rawToken string) (string, error) {
	tokenHash := hashToken(rawToken)

	var userID string
	var expiresAt time.Time
	var usedAt sql.NullTime
	err := db.QueryRowContext(
		ctx,
		"SELECT user_id, expires_at, used_at FROM password_reset_tokens WHERE token_hash = $1",
		tokenHash,
	).Scan(&userID, &expiresAt, &usedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrPasswordResetTokenNotFound
		}
		return "", err
	}

	if usedAt.Valid {
		return "", ErrPasswordResetTokenUsed
	}

	if time.Now().After(expiresAt) {
		return "", ErrPasswordResetTokenExpired
	}

	return userID, nil
}

type execContexter interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func ConsumePasswordResetToken(ctx context.Context, db execContexter, rawToken string) error {
	tokenHash := hashToken(rawToken)
	result, err := db.ExecContext(
		ctx,
		"UPDATE password_reset_tokens SET used_at = NOW() WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()",
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
		return ErrPasswordResetTokenNotFound
	}
	return nil
}
