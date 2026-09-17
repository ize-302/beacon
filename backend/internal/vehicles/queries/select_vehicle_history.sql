SELECT
	v.id,
	v.plate_number,
	lp.latitude,
	lp.longitude
FROM vehicles v
LEFT JOIN LATERAL (
	SELECT latitude, longitude
	FROM gpspoints
	WHERE vehicle_id = v.id
	ORDER BY created_at DESC
	LIMIT 200
) lp ON true
WHERE v.id = $1
