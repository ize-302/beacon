-- Inserts a whole batch of gpspoints in one statement. unnest() zips the five parallel arrays
-- back into rows, so the parameter count stays at 5 no matter how many points
-- are sent
INSERT INTO gpspoints (gps_id, bearing, latitude, longitude, timestamp)
SELECT * FROM unnest(
	$1::integer[],
	$2::double precision[],
	$3::double precision[],
	$4::double precision[],
	$5::bigint[]
)
