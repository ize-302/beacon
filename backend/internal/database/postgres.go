// Package database
package database

import (
	"database/sql"
)

type Handler struct {
	DB *sql.DB
}

func DBConn(conn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}

	return db, db.Ping()
}
