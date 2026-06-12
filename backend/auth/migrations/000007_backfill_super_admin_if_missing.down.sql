UPDATE auth
SET role = 'free_user'
WHERE role = 'super_admin'
  AND user_id IN (
      SELECT user_id
      FROM auth
      ORDER BY created_at ASC, id ASC
      LIMIT 1
  )
  AND (SELECT COUNT(*) FROM auth WHERE role = 'super_admin') = 1;
