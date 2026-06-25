package services

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

const ClientInvitationTokenTTL = 72 * time.Hour

var (
	ErrClientInvitationTokenNotFound = errors.New("client invitation token not found")
	ErrClientInvitationTokenExpired  = errors.New("client invitation token expired")
	ErrClientInvitationTokenUsed     = errors.New("client invitation token already used")
)

func GenerateClientInvitationToken(ctx context.Context, db *sql.DB, clientID, providerID string) (string, error) {
	rawToken, err := generateRawToken()
	if err != nil {
		return "", err
	}

	tokenHash := hashToken(rawToken)
	expiresAt := time.Now().Add(ClientInvitationTokenTTL)

	_, err = db.ExecContext(
		ctx,
		"INSERT INTO client_invitation_tokens (client_id, provider_id, token_hash, expires_at) VALUES ($1, $2, $3, $4)",
		clientID,
		providerID,
		tokenHash,
		expiresAt,
	)
	if err != nil {
		return "", err
	}

	return rawToken, nil
}

func ValidateClientInvitationToken(ctx context.Context, db *sql.DB, rawToken string) (clientID, providerID string, err error) {
	tokenHash := hashToken(rawToken)

	var expiresAt time.Time
	var usedAt sql.NullTime
	err = db.QueryRowContext(
		ctx,
		"SELECT client_id, provider_id, expires_at, used_at FROM client_invitation_tokens WHERE token_hash = $1",
		tokenHash,
	).Scan(&clientID, &providerID, &expiresAt, &usedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", ErrClientInvitationTokenNotFound
		}
		return "", "", err
	}

	if usedAt.Valid {
		return "", "", ErrClientInvitationTokenUsed
	}

	if time.Now().After(expiresAt) {
		return "", "", ErrClientInvitationTokenExpired
	}

	return clientID, providerID, nil
}

func ConsumeClientInvitationToken(ctx context.Context, db execContexter, rawToken string) error {
	tokenHash := hashToken(rawToken)
	result, err := db.ExecContext(
		ctx,
		"UPDATE client_invitation_tokens SET used_at = NOW() WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()",
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
		return ErrClientInvitationTokenNotFound
	}
	return nil
}
