WITH inserted AS (
	INSERT INTO locations (assignment_id, latitude, longitude)
	VALUES ($1, $2, $3)
	RETURNING id, assignment_id, latitude, longitude, created_at 
)
SELECT
	i.id AS location_id,
	i.latitude AS location_latitude,
	i.longitude AS location_longitude,
	i.created_at AS location_created_at,
	a.id AS assignment_id,
	a.created_at,
	v.id AS vehicle_id, 
	v.plate_number AS vehicle_platenumber, 
	v.created_at AS vehicle_created_at, 
	r.id AS rider_id, 
	r.name AS rider_name, 
	r.created_at AS rider_created_at
FROM inserted i
JOIN assignments a ON a.id = i.assignment_id
INNER JOIN vehicles v ON a.vehicle_id = v.id
INNER JOIN riders r ON a.rider_id = r.id


