package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"
)

const PasswordResetTokenTTL = 15 * time.Minute

var (
	ErrPasswordResetTokenNotFound = errors.New("password reset token not found")
	ErrPasswordResetTokenExpired  = errors.New("password reset token expired")
	ErrPasswordResetTokenUsed     = errors.New("password reset token already used")
)

func hashPasswordResetToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func generatePasswordResetToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func GeneratePasswordResetToken(ctx context.Context, db *sql.DB, userID string) (string, error) {
	rawToken, err := generatePasswordResetToken()
	if err != nil {
		return "", err
	}

	tokenHash := hashPasswordResetToken(rawToken)
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
	tokenHash := hashPasswordResetToken(rawToken)

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

func ConsumePasswordResetToken(ctx context.Context, db *sql.DB, rawToken string) error {
	return consumePasswordResetTokenWithExecer(ctx, db, rawToken)
}

func ConsumePasswordResetTokenTx(ctx context.Context, tx *sql.Tx, rawToken string) error {
	return consumePasswordResetTokenWithExecer(ctx, tx, rawToken)
}

type execContexter interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func consumePasswordResetTokenWithExecer(ctx context.Context, execer execContexter, rawToken string) error {
	tokenHash := hashPasswordResetToken(rawToken)

	result, err := execer.ExecContext(
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
