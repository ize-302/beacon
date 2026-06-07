INSERT INTO riders (name) VALUES ($1) RETURNING id, name, created_at
