ALTER TABLE users
  ADD COLUMN first_name TEXT,
  ADD COLUMN last_name TEXT,
  ADD COLUMN role TEXT NOT NULL DEFAULT 'user',
  ADD COLUMN plan TEXT NOT NULL DEFAULT 'starter',
  ADD COLUMN last_login_at TIMESTAMP,
  ADD COLUMN suspended BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE users
  ADD CONSTRAINT users_role_check CHECK (role IN ('user', 'admin'));
