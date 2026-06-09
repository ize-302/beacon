WITH inserted AS (
	INSERT INTO gpspoints (gps_id, latitude, longitude)
	VALUES ($1, $2, $3)
	RETURNING id, gps_id, latitude, longitude, created_at 
)
SELECT
	gpsp.id AS gpspoint_id,
	gps.id AS gps_id, 
	gpsp.latitude AS gpspoint_latitude,
	gpsp.longitude AS gpspoint_longitude,
	gpsp.created_at AS gpspoint_created_at
FROM inserted gpsp
INNER JOIN gps gps ON gpsp.gps_id = gps.id


