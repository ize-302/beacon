// Package database
package database

import (
	_ "embed"
)

//go:embed queries/create_vehicles_table.sql
var createVehiclesTable string

//go:embed queries/create_gps_table.sql
var createGpsTable string

//go:embed queries/create_gpspoints_table.sql
var createGpsPointsTable string

func (h *Handler) SeedDB() error {
	tx, err := h.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// create Vehicle table
	_, err = tx.Exec(createVehiclesTable)
	if err != nil {
		return err
	}

	// create Vehicle table
	_, err = tx.Exec(createGpsTable)
	if err != nil {
		return err
	}

	// create gpspoints table
	_, err = tx.Exec(createGpsPointsTable)
	if err != nil {
		return err
	}

	return tx.Commit()
}
