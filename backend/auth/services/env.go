package services

import (
	"os"
)

type EnvKey string

const (
 Env                     EnvKey = "ENV"
 APPHost                 EnvKey = "API_HOST"
 APPPort                 EnvKey = "API_PORT"
 APPKey					 EnvKey = "API_KEY"
 PostgresUser            EnvKey = "POSTGRES_USER"
 PostgresPassword        EnvKey = "POSTGRES_PASSWORD"
 PostgresDefaultDatabase EnvKey = "POSTGRES_DB"
 PostgresDatabaseAddress EnvKey = "POSTGRES_DB_ADDRESS"
 PostgresDatabasePort    EnvKey = "POSTGRES_DB_PORT"
)

func (key EnvKey) GetValue() string {
 return os.Getenv(string(key))
}