SELECT
	v.id,
	v.plate_number,
	v.vehicle_type,
	COALESCE(v.device_sn, ''),
	v.created_at,
	lp.latitude,
	lp.longitude,
	lp.created_at AS last_point_at
FROM vehicles v
LEFT JOIN LATERAL (
	SELECT latitude, longitude, created_at
	FROM gpspoints
	WHERE vehicle_id = v.id
	ORDER BY created_at DESC
	LIMIT 1
) lp ON true
ORDER BY v.id
