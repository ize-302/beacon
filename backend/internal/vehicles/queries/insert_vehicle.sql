INSERT INTO vehicles (plate_number) VALUES ($1) RETURNING id, plate_number, created_at;
