SELECT 
	a.id AS assignment_id,
	v.id AS vehicle_id, 
	v.plate_number, 
	r.id AS rider_id, 
	r.name AS rider_name
FROM assignments a
INNER JOIN vehicles v ON a.vehicle_id = v.id
INNER JOIN riders r ON a.rider_id = r.id

