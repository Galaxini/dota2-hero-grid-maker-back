DELETE FROM grids
WHERE user_id IS NULL
  AND title = 'Default Grid';
