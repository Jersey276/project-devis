package services

import "os"

type EnvKey string

const (
	DBHost         EnvKey = "DB_HOST"
	DBPort         EnvKey = "DB_PORT"
	DBUser         EnvKey = "DB_USER"
	DBPassword     EnvKey = "DB_PASSWORD"
	DBPasswordFile EnvKey = "DB_PASSWORD_FILE"
	DBName         EnvKey = "DB_NAME"

	PurgeDBUser         EnvKey = "PURGE_DB_USER"
	PurgeDBPasswordFile EnvKey = "PURGE_DB_PASSWORD_FILE"
	PurgeDBName         EnvKey = "PURGE_DB_NAME"

	EmailServiceAddress EnvKey = "EMAIL_SERVICE_ADDRESS"
)

func (key EnvKey) GetValue() string {
	return os.Getenv(string(key))
}
