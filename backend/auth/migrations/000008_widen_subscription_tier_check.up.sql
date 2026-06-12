ALTER TABLE auth DROP CONSTRAINT auth_subscription_tier_check;
ALTER TABLE auth ADD CONSTRAINT auth_subscription_tier_check
    CHECK (subscription_tier IN ('free', 'pro', 'enterprise'));
