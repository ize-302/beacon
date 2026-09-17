INSERT INTO gpspoints (vehicle_id, bearing, latitude, longitude, timestamp)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, vehicle_id, bearing, latitude, longitude, timestamp, created_at
