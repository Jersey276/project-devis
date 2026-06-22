package services

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func connectDB(user, passwordFile, name string) *sql.DB {
	host := DBHost.GetValue()
	if host == "" {
		host = "localhost"
	}
	port := DBPort.GetValue()
	if port == "" {
		port = "5432"
	}

	password := ""
	if path := passwordFile; path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("failed to read password file %s: %v", path, err)
		}
		password = strings.TrimSpace(string(data))
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, name,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	return db
}

func ConnectDB() *sql.DB {
	return connectDB(
		DBUser.GetValue(),
		DBPasswordFile.GetValue(),
		DBName.GetValue(),
	)
}

// ConnectPurgeDB returns a connection for the purge role (DELETE-only).
func ConnectPurgeDB() *sql.DB {
	user := PurgeDBUser.GetValue()
	if user == "" {
		return nil
	}
	return connectDB(
		user,
		PurgeDBPasswordFile.GetValue(),
		PurgeDBName.GetValue(),
	)
}
