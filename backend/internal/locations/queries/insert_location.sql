WITH inserted AS (
	INSERT INTO locations (vehicle_id, latitude, longitude)
	VALUES ($1, $2, $3)
	RETURNING id, vehicle_id, latitude, longitude, created_at 
)
SELECT
	i.id,
	i.latitude,
	i.longitude,
	i.created_at,
	v.id,
	v.plate_number,
	v.created_at
FROM inserted i
JOIN vehicles v ON v.id = i.vehicle_id

