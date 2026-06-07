// Package database
package database

import (
	_ "embed"
)

//go:embed queries/create_vehicles_table.sql
var createVehiclesTable string

//go:embed queries/create_riders_table.sql
var createRidersTable string

//go:embed queries/create_assignments_table.sql
var createAssignmentsTable string

//go:embed queries/create_locations_table.sql
var createLocationsTable string

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

	// create riders table
	_, err = tx.Exec(createRidersTable)
	if err != nil {
		return err
	}

	// create assignments table
	_, err = tx.Exec(createAssignmentsTable)
	if err != nil {
		return err
	}

	// create locations table
	_, err = tx.Exec(createLocationsTable)
	if err != nil {
		return err
	}

	return tx.Commit()
}
