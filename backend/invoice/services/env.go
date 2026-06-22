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
	QuoteServiceAddress     EnvKey = "QUOTE_SERVICE_ADDRESS"
	UserServiceAddress      EnvKey = "USER_SERVICE_ADDRESS"
	ScheduleServiceAddress  EnvKey = "SCHEDULE_SERVICE_ADDRESS"
	ExportServiceAddress    EnvKey = "EXPORT_SERVICE_ADDRESS"
	PDPServiceAddress       EnvKey = "PDP_SERVICE_ADDRESS"
	// PA (Plateforme Agréée) adapter selection — see pdp/ and pdp/iopole/.
	PDPProvider        EnvKey = "PDP_PROVIDER" // "noop" (default) | "iopole"
	IopoleBaseURL      EnvKey = "IOPOLE_BASE_URL"
	IopoleTokenURL     EnvKey = "IOPOLE_TOKEN_URL"
	IopoleClientID     EnvKey = "IOPOLE_CLIENT_ID"
	IopoleClientSecret EnvKey = "IOPOLE_CLIENT_SECRET"
	IopoleCustomerID   EnvKey = "IOPOLE_CUSTOMER_ID"
)

func (key EnvKey) GetValue() string {
	return os.Getenv(string(key))
}
