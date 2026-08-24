// Package database
package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Handler struct {
	DB *sql.DB
}

// use connString in deployment platforms like railwway
// fall back to variables locally
func connString() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)
}

func DBConn() (*sql.DB, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	db, err := sql.Open("postgres", connString())
	if err != nil {
		return nil, err
	}

	// private network on some platforms don't resolve the instant the
	// container starts, so give the first connection a few attempts
	for i := range 10 {
		if err = db.Ping(); err == nil {
			return db, nil
		}
		log.Printf("database not reachable (attempt %d/10): %v", i+1, err)
		time.Sleep(time.Second)
	}

	db.Close()
	return nil, err
}
