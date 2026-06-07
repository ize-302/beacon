	WITH inserted AS (
			INSERT INTO assignments (vehicle_id, rider_id)
			VALUES ($1, $2)
			RETURNING id, vehicle_id, rider_id
		)
		SELECT
			i.id,
			v.id,
			v.plate_number,
			r.id,
			r.name
		FROM inserted i
		JOIN vehicles v ON v.id = i.vehicle_id
		JOIN riders r ON r.id = i.rider_id

