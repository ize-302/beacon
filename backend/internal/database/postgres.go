// Package database
package database

import (
	"database/sql"
)

func DBConn(conn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}

	return db, db.Ping()
}
