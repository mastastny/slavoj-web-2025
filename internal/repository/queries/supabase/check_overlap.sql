SELECT COUNT(*)
FROM reservations
WHERE court_id = $1
  AND start_at < $2
  AND end_at   > $3
