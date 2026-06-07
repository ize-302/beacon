// Package database
package database

func (h *Handler) SeedDB() error {
	tx, err := h.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// create Vehicle table
	createVehiclesTable := `CREATE TABLE IF NOT EXISTS vehicles (
		id SERIAL PRIMARY KEY,
		plate_number TEXT NOT NULL UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	_, err = tx.Exec(createVehiclesTable)
	if err != nil {
		return err
	}

	// create riders table
	createRidersTable := `CREATE TABLE IF NOT EXISTS riders (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	_, err = tx.Exec(createRidersTable)
	if err != nil {
		return err
	}

	// create assignments table
	createAssignmentsTable := `CREATE TABLE IF NOT EXISTS assignments (
		id SERIAL PRIMARY KEY,
		vehicle_id INTEGER NOT NULL UNIQUE,
		rider_id INTEGER NOT NULL UNIQUE,
 		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	_, err = tx.Exec(createAssignmentsTable)
	if err != nil {
		return err
	}

	// create locations table
	createLocationsTable := `CREATE TABLE IF NOT EXISTS locations (
		id SERIAL PRIMARY KEY,
		latitude FLOAT NOT NULL,
		longitude FLOAT NOT NULL,
		vehicle_id INTEGER NOT NULL,
 		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	_, err = tx.Exec(createLocationsTable)
	if err != nil {
		return err
	}

	return tx.Commit()
}
