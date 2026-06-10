SELECT
	g.id AS gps_id,
	g.sn AS gps_sn,
	g.created_at,
	v.id AS vehicle_id,
	v.plate_number AS vehicle_plate_number,
	v.created_at AS vehicle_created_at,
	lp.latitude,
	lp.longitude,
	lp.created_at AS last_point_at
FROM gps g
LEFT JOIN vehicles v ON g.vehicle_id = v.id
LEFT JOIN LATERAL (
	SELECT latitude, longitude, created_at
	FROM gpspoints
	WHERE gps_id = g.id
	ORDER BY created_at DESC
	LIMIT 1
) lp ON true
