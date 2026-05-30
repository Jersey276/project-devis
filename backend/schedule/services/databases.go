package services

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func ConnectDB() *sql.DB {
	host := DBHost.GetValue()
	if host == "" {
		host = PostgresDatabaseAddress.GetValue()
	}
	if host == "" {
		host = "localhost"
	}

	port := DBPort.GetValue()
	if port == "" {
		port = PostgresDatabasePort.GetValue()
	}
	if port == "" {
		port = "5432"
	}

	user := DBUser.GetValue()
	if user == "" {
		user = PostgresUser.GetValue()
	}

	password := DBPassword.GetValue()
	if password == "" {
		if path := DBPasswordFile.GetValue(); path != "" {
			data, err := os.ReadFile(path)
			if err != nil {
				log.Fatalf("failed to read DB_PASSWORD_FILE: %v", err)
			}
			password = strings.TrimSpace(string(data))
		}
	}
	if password == "" {
		password = PostgresPassword.GetValue()
	}

	name := DBName.GetValue()
	if name == "" {
		name = PostgresDefaultDatabase.GetValue()
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