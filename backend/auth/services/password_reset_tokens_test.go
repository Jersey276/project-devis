package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGeneratePasswordResetToken_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(`INSERT INTO password_reset_tokens \(user_id, token_hash, expires_at\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs("user-123", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	token, err := GeneratePasswordResetToken(context.Background(), db, "user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected a non-empty reset token")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestValidatePasswordResetToken_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	future := time.Now().Add(10 * time.Minute)
	mock.ExpectQuery(`SELECT user_id, expires_at, used_at FROM password_reset_tokens WHERE token_hash = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at", "used_at"}).AddRow("user-123", future, nil))

	userID, err := ValidatePasswordResetToken(context.Background(), db, "raw-reset-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != "user-123" {
		t.Fatalf("expected user-123, got %s", userID)
	}
}

func TestValidatePasswordResetToken_Expired(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	past := time.Now().Add(-1 * time.Minute)
	mock.ExpectQuery(`SELECT user_id, expires_at, used_at FROM password_reset_tokens WHERE token_hash = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at", "used_at"}).AddRow("user-123", past, nil))

	_, err = ValidatePasswordResetToken(context.Background(), db, "raw-reset-token")
	if !errors.Is(err, ErrPasswordResetTokenExpired) {
		t.Fatalf("expected ErrPasswordResetTokenExpired, got %v", err)
	}
}

func TestValidatePasswordResetToken_Used(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	future := time.Now().Add(10 * time.Minute)
	now := time.Now()
	mock.ExpectQuery(`SELECT user_id, expires_at, used_at FROM password_reset_tokens WHERE token_hash = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at", "used_at"}).AddRow("user-123", future, now))

	_, err = ValidatePasswordResetToken(context.Background(), db, "raw-reset-token")
	if !errors.Is(err, ErrPasswordResetTokenUsed) {
		t.Fatalf("expected ErrPasswordResetTokenUsed, got %v", err)
	}
}

func TestConsumePasswordResetToken_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(`UPDATE password_reset_tokens SET used_at = NOW\(\) WHERE token_hash = \$1 AND used_at IS NULL AND expires_at > NOW\(\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = ConsumePasswordResetToken(context.Background(), db, "raw-reset-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConsumePasswordResetToken_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(`UPDATE password_reset_tokens SET used_at = NOW\(\) WHERE token_hash = \$1 AND used_at IS NULL AND expires_at > NOW\(\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = ConsumePasswordResetToken(context.Background(), db, "raw-reset-token")
	if !errors.Is(err, ErrPasswordResetTokenNotFound) {
		t.Fatalf("expected ErrPasswordResetTokenNotFound, got %v", err)
	}
}
