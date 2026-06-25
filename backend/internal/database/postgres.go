// Package database
package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Handler struct {
	DB *sql.DB
}

func DBConn() (*sql.DB, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	conn := fmt.Sprintf("host=%s port=%s user=%s  password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}

	return db, db.Ping()
}
