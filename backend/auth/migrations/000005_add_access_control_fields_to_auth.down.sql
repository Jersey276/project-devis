ALTER TABLE auth
    DROP CONSTRAINT IF EXISTS auth_subscription_tier_check,
    DROP CONSTRAINT IF EXISTS auth_account_status_check,
    DROP CONSTRAINT IF EXISTS auth_role_check,
    DROP COLUMN IF EXISTS subscription_tier,
    DROP COLUMN IF EXISTS account_status,
    DROP COLUMN IF EXISTS role;
