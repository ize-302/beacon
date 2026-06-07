SELECT 
	l.id AS location_id,
	l.latitude AS location_latitude,
	l.longitude AS location_longitude,
	l.created_at AS location_created_at,
	v.id AS vehicle_id, 
	v.plate_number, 
	v.created_at
FROM locations l
INNER JOIN vehicles v ON l.vehicle_id = v.id

