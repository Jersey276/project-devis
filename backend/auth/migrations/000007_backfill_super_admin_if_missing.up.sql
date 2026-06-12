WITH has_super_admin AS (
    SELECT EXISTS (
        SELECT 1 FROM auth WHERE role = 'super_admin'
    ) AS exists_admin
), oldest_account AS (
    SELECT user_id
    FROM auth
    ORDER BY created_at ASC, id ASC
    LIMIT 1
)
UPDATE auth a
SET role = 'super_admin'
FROM has_super_admin h, oldest_account o
WHERE h.exists_admin = false
  AND a.user_id = o.user_id;
