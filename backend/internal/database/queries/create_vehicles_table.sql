CREATE TABLE IF NOT EXISTS vehicles (
	id SERIAL PRIMARY KEY,
	plate_number TEXT NOT NULL UNIQUE,
	vehicle_type TEXT NOT NULL CHECK (vehicle_type IN ('car', 'bus', 'truck', 'van', 'tricycle', 'motorcycle')),
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
