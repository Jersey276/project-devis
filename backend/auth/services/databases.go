package services

import (
	"database/sql"
	"fmt"
	"log"

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
		password = PostgresPassword.GetValue()
	}

	name := DBName.GetValue()
	if name == "" {
		name = PostgresDefaultDatabase.GetValue()
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		user,
		password,
		name,
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