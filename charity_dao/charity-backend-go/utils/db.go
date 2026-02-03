package utils

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
	dsn := "postgres://root:1234@localhost:5432/charity_dao?sslmode=disable"

	db, err := sql.Open("postgres", dsn)

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	DB = db
	log.Println("Connected Database")
}
