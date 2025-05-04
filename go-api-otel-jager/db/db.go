package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	// "github.com/uptrace/opentelemetry-go-extra/otelsql"
)

var DB *sql.DB

func InitDB() {
	var err error
	// connStr := "postgres://admin:mysecretpassword@localhost:5432/zaq?sslmode=disable"
	connStr := "postgres://admin:mysecretpassword@postgres:5432/zaq?sslmode=disable"
	// Adjust for your environment
	DB, err = sql.Open("postgres", connStr)

	// Open the database with OpenTelemetry instrumentation
	// DB, err = otelsql.Open("postgres", connStr, otelsql.WithAttributes(
	// 	semconv.DBSystemPostgreSQL,
	// ))
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
		return
	}

	// Check if DB is nil
	if DB == nil {
		log.Fatal("Database connection is nil after initialization")
		return
	}

	fmt.Println(" Database Connected!")
	DB.SetMaxOpenConns(10)
	createTables()

}

func createTables() {
	createUsersTable := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			password TEXT NOT NULL
		)
	`
	_, err := DB.Exec(createUsersTable)
	if err != nil {
		log.Fatal("Couldn't create users table:", err.Error())
	}

	createEventsTable := `
		CREATE TABLE IF NOT EXISTS events (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			location TEXT NOT NULL,
			dateTime TIMESTAMP NOT NULL,
			user_id INTEGER,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`
	_, err = DB.Exec(createEventsTable)
	if err != nil {
		log.Fatal("Couldn't create events table:", err.Error())
	}

	createRegistrationTable := `
		CREATE TABLE IF NOT EXISTS registrations (
			id SERIAL PRIMARY KEY,
			event_id INTEGER,
			user_id INTEGER,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
		)
	`
	_, err = DB.Exec(createRegistrationTable)
	if err != nil {
		log.Fatal("Couldn't create registrations table:", err.Error())
	}
}
