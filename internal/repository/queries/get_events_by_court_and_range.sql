SELECT name, start_at, end_at
FROM reservations
WHERE court_id = ?
  AND start_at >= ?
  AND end_at   <= ?
ORDER BY start_at
