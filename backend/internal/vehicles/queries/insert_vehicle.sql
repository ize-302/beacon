INSERT INTO vehicles (plate_number, vehicle_type, device_sn)
VALUES ($1, $2, NULLIF($3, ''))
RETURNING id, plate_number, vehicle_type, COALESCE(device_sn, ''), created_at
