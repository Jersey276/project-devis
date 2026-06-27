package services

import "os"

type EnvKey string

const (
	Env                     EnvKey = "ENV"
	PostgresUser            EnvKey = "POSTGRES_USER"
	PostgresPassword        EnvKey = "POSTGRES_PASSWORD"
	PostgresDefaultDatabase EnvKey = "POSTGRES_DB"
	PostgresDatabaseAddress EnvKey = "POSTGRES_DB_ADDRESS"
	PostgresDatabasePort    EnvKey = "POSTGRES_DB_PORT"
	DBHost                  EnvKey = "DB_HOST"
	DBPort                  EnvKey = "DB_PORT"
	DBUser                  EnvKey = "DB_USER"
	DBPassword              EnvKey = "DB_PASSWORD"
	DBPasswordFile          EnvKey = "DB_PASSWORD_FILE"
	DBName                  EnvKey = "DB_NAME"
	ResendAPIKey     EnvKey = "RESEND_API_KEY"
	ResendAPIKeyFile EnvKey = "RESEND_API_KEY_FILE"
	EmailFrom        EnvKey = "EMAIL_FROM"
	EmailFromFile    EnvKey = "EMAIL_FROM_FILE"
)

func (key EnvKey) GetValue() string {
	return os.Getenv(string(key))
}
