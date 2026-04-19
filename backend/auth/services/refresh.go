package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
)

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func GenerateRefreshToken(ctx context.Context, db *sql.DB, userID string, rememberMe bool) (string, error) {
	raw := uuid.New().String()
	hash := hashToken(raw)

	expiry := 7 * 24 * time.Hour
	if rememberMe {
		expiry = 60 * 24 * time.Hour // ~2 months
	}
	expiresAt := time.Now().Add(expiry)

	_, err := db.ExecContext(ctx,
		"INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)",
		userID, hash, expiresAt,
	)
	if err != nil {
		return "", err
	}

	return raw, nil
}

func ValidateRefreshToken(ctx context.Context, db *sql.DB, rawToken string) (string, error) {
	hash := hashToken(rawToken)

	var userID string
	var expiresAt time.Time
	err := db.QueryRowContext(ctx,
		"SELECT user_id, expires_at FROM refresh_tokens WHERE token_hash = $1",
		hash,
	).Scan(&userID, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrRefreshTokenNotFound
		}
		return "", err
	}

	if time.Now().After(expiresAt) {
		// Clean up expired token
		_, _ = db.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE token_hash = $1", hash)
		return "", ErrRefreshTokenExpired
	}

	return userID, nil
}

func DeleteRefreshToken(ctx context.Context, db *sql.DB, rawToken string) error {
	hash := hashToken(rawToken)
	_, err := db.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE token_hash = $1", hash)
	return err
}

func DeleteAllRefreshTokens(ctx context.Context, db *sql.DB, userID string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE user_id = $1", userID)
	return err
}
