INSERT INTO reservations (court_id, start_at, end_at, name, email)
VALUES ($1, $2, $3, $4, $5)
RETURNING id
