ALTER TABLE auth
    ADD COLUMN role TEXT NOT NULL DEFAULT 'free_user',
    ADD COLUMN account_status TEXT NOT NULL DEFAULT 'active',
    ADD COLUMN subscription_tier TEXT NOT NULL DEFAULT 'free';

ALTER TABLE auth
    ADD CONSTRAINT auth_role_check CHECK (role IN ('free_user', 'super_admin')),
    ADD CONSTRAINT auth_account_status_check CHECK (account_status IN ('active', 'suspended')),
    ADD CONSTRAINT auth_subscription_tier_check CHECK (subscription_tier IN ('free'));
