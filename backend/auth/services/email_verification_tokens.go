package services

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

const EmailVerificationTokenTTL = 24 * time.Hour

var (
	ErrEmailVerificationTokenNotFound = errors.New("email verification token not found")
	ErrEmailVerificationTokenExpired  = errors.New("email verification token expired")
	ErrEmailVerificationTokenUsed     = errors.New("email verification token already used")
)

func GenerateEmailVerificationToken(ctx context.Context, db *sql.DB, userID string) (string, error) {
	rawToken, err := generateRawToken()
	if err != nil {
		return "", err
	}
	tokenHash := hashToken(rawToken)
	expiresAt := time.Now().Add(EmailVerificationTokenTTL)

	_, err = db.ExecContext(ctx,
		"INSERT INTO email_verification_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)",
		userID, tokenHash, expiresAt,
	)
	if err != nil {
		return "", err
	}
	return rawToken, nil
}

func ValidateEmailVerificationToken(ctx context.Context, db *sql.DB, rawToken string) (string, error) {
	tokenHash := hashToken(rawToken)

	var userID string
	var expiresAt time.Time
	var usedAt sql.NullTime
	err := db.QueryRowContext(ctx,
		"SELECT user_id, expires_at, used_at FROM email_verification_tokens WHERE token_hash = $1",
		tokenHash,
	).Scan(&userID, &expiresAt, &usedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrEmailVerificationTokenNotFound
		}
		return "", err
	}

	if usedAt.Valid {
		return "", ErrEmailVerificationTokenUsed
	}
	if time.Now().After(expiresAt) {
		return "", ErrEmailVerificationTokenExpired
	}
	return userID, nil
}

func ConsumeEmailVerificationToken(ctx context.Context, db *sql.DB, rawToken string) error {
	tokenHash := hashToken(rawToken)
	result, err := db.ExecContext(ctx,
		"UPDATE email_verification_tokens SET used_at = NOW() WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()",
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
		return ErrEmailVerificationTokenNotFound
	}
	return nil
}
