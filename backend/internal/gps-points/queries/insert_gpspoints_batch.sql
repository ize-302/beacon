INSERT INTO gpspoints (vehicle_id, bearing, latitude, longitude, timestamp)
SELECT * FROM unnest(
	$1::integer[],
	$2::double precision[],
	$3::double precision[],
	$4::double precision[],
	$5::bigint[]
)
