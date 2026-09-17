CREATE INDEX IF NOT EXISTS idx_gpspoints_vehicle_recent
	ON gpspoints (vehicle_id, created_at DESC)
