SELECT 
	g.id AS gps_id, 
	g.sn AS gps_sn,
	g.created_at,
	v.id AS vehicle_id,
	v.plate_number AS vehicle_plate_number,
	v.created_at AS vehicle_created_at
FROM gps g
LEFT JOIN vehicles v ON g.vehicle_id = v.id
