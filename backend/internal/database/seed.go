// Package database
package database

import (
	_ "embed"
)

//go:embed queries/create_vehicles_table.sql
var createVehiclesTable string

//go:embed queries/create_gpspoints_table.sql
var createGpsPointsTable string

//go:embed queries/create_gpspoints_index.sql
var createGpsPointsIndex string

func (h *Handler) SeedDB() error {
	tx, err := h.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, stmt := range []string{
		createVehiclesTable,
		createGpsPointsTable,
		createGpsPointsIndex,
	} {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}

	return tx.Commit()
}
