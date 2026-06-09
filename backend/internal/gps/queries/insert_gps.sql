WITH inserted AS (
	INSERT INTO gps (sn, vehicle_id)
	VALUES ($1, $2)
	RETURNING id, sn, vehicle_id, created_at
)
SELECT 
	g.id AS gps_id, 
	g.sn AS gps_sn,
	g.created_at,
	v.id AS vehicle_id,
	v.plate_number AS vehicle_plate_number,
	v.created_at AS vehicle_created_at
FROM inserted g
LEFT JOIN vehicles v ON g.vehicle_id = v.id
