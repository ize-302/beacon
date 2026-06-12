INSERT INTO vehicles (plate_number, vehicle_type) VALUES ($1, $2) RETURNING id, plate_number, vehicle_type, created_at;
