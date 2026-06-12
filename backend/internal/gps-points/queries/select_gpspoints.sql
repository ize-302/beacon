SELECT 
	gpsp.id AS gpspoint_id,
	gps.id AS gps_id,
	gpsp.latitude AS gpspoint_latitude,
	gpsp.longitude AS gpspoint_longitude,
	gpsp.timestamp AS gpspoint_timestamp,
	gpsp.created_at AS gpspoint_created_at
	FROM gpspoints gpsp
INNER JOIN gps_devices gps ON gpsp.gps_id = gps.id

