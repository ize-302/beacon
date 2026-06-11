SELECT
	g.id AS gps_id,
	g.sn AS gps_sn,
	lp.latitude AS gpspoint_latitude,
	lp.longitude AS gpspoint_longitude
FROM gps g
LEFT JOIN LATERAL (
	SELECT latitude, longitude
	FROM gpspoints
	WHERE gps_id = g.id
	ORDER BY created_at DESC
	LIMIT 200
) lp ON true
WHERE g.id = $1
