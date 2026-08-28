-- gpspoints.vehicle_id is ON DELETE CASCADE, so the database removes the
-- vehicle's history for us.
DELETE FROM vehicles WHERE id = $1
