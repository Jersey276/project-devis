package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGenerateEmailVerificationToken_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(`INSERT INTO email_verification_tokens \(user_id, token_hash, expires_at\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs("user-abc", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	token, err := GenerateEmailVerificationToken(context.Background(), db, "user-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected a non-empty verification token")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestValidateEmailVerificationToken_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	future := time.Now().Add(10 * time.Minute)
	mock.ExpectQuery(`SELECT user_id, expires_at, used_at FROM email_verification_tokens WHERE token_hash = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at", "used_at"}).AddRow("user-abc", future, nil))

	userID, err := ValidateEmailVerificationToken(context.Background(), db, "raw-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != "user-abc" {
		t.Fatalf("expected user-abc, got %s", userID)
	}
}

func TestValidateEmailVerificationToken_Expired(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	past := time.Now().Add(-1 * time.Minute)
	mock.ExpectQuery(`SELECT user_id, expires_at, used_at FROM email_verification_tokens WHERE token_hash = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at", "used_at"}).AddRow("user-abc", past, nil))

	_, err = ValidateEmailVerificationToken(context.Background(), db, "raw-token")
	if !errors.Is(err, ErrEmailVerificationTokenExpired) {
		t.Fatalf("expected ErrEmailVerificationTokenExpired, got %v", err)
	}
}

func TestValidateEmailVerificationToken_Used(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	future := time.Now().Add(10 * time.Minute)
	usedAt := time.Now()
	mock.ExpectQuery(`SELECT user_id, expires_at, used_at FROM email_verification_tokens WHERE token_hash = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at", "used_at"}).AddRow("user-abc", future, usedAt))

	_, err = ValidateEmailVerificationToken(context.Background(), db, "raw-token")
	if !errors.Is(err, ErrEmailVerificationTokenUsed) {
		t.Fatalf("expected ErrEmailVerificationTokenUsed, got %v", err)
	}
}

func TestValidateEmailVerificationToken_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT user_id, expires_at, used_at FROM email_verification_tokens WHERE token_hash = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at", "used_at"}))

	_, err = ValidateEmailVerificationToken(context.Background(), db, "unknown-token")
	if !errors.Is(err, ErrEmailVerificationTokenNotFound) {
		t.Fatalf("expected ErrEmailVerificationTokenNotFound, got %v", err)
	}
}

func TestConsumeEmailVerificationToken_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(`UPDATE email_verification_tokens SET used_at = NOW\(\) WHERE token_hash = \$1 AND used_at IS NULL AND expires_at > NOW\(\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := ConsumeEmailVerificationToken(context.Background(), db, "raw-token"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConsumeEmailVerificationToken_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(`UPDATE email_verification_tokens SET used_at = NOW\(\) WHERE token_hash = \$1 AND used_at IS NULL AND expires_at > NOW\(\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = ConsumeEmailVerificationToken(context.Background(), db, "raw-token")
	if !errors.Is(err, ErrEmailVerificationTokenNotFound) {
		t.Fatalf("expected ErrEmailVerificationTokenNotFound, got %v", err)
	}
}
