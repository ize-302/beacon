WITH inserted AS (
	INSERT INTO gpspoints (gps_id, bearing, latitude, longitude, timestamp)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, gps_id, bearing, latitude, longitude, timestamp, created_at 
)
SELECT
	gpsp.id AS gpspoint_id,
	gps.id AS gps_id, 
	gpsp.bearing AS gpspoint_bearing, 
	gpsp.latitude AS gpspoint_latitude,
	gpsp.longitude AS gpspoint_longitude,
	gpsp.timestamp AS gpspoint_timestamp,
	gpsp.created_at AS gpspoint_created_at
FROM inserted gpsp
INNER JOIN gps_devices gps ON gpsp.gps_id = gps.id


