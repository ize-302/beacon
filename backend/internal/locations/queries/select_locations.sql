SELECT 
	l.id AS location_id,
	l.latitude AS location_latitude,
	l.longitude AS location_longitude,
	l.created_at AS location_created_at,
	a.id AS assignment_id,
	a.created_at,
	v.id AS vehicle_id, 
	v.plate_number AS vehicle_platenumber, 
	v.created_at AS vehicle_created_at, 
	r.id AS rider_id, 
	r.name AS rider_name, 
	r.created_at AS rider_created_at
FROM locations l
INNER JOIN assignments a ON l.assignment_id = a.id
INNER JOIN vehicles v ON a.vehicle_id = v.id
INNER JOIN riders r ON a.rider_id = r.id

