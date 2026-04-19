package services

import (
	"os"
)

type EnvKey string

const (
 Env                     EnvKey = "ENV"
 APPHost                 EnvKey = "API_HOST"
 APPPort                 EnvKey = "API_PORT"
 APPKey                  EnvKey = "APP_KEY"
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
 UserServiceAddress      EnvKey = "USER_SERVICE_ADDRESS"
)

func (key EnvKey) GetValue() string {
 return os.Getenv(string(key))
}