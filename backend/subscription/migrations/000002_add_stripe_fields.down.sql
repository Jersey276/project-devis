DROP TABLE IF EXISTS subscription_events;

ALTER TABLE subscriptions
    DROP COLUMN IF EXISTS cancel_at_period_end,
    DROP COLUMN IF EXISTS stripe_subscription_id,
    DROP COLUMN IF EXISTS stripe_customer_id;

ALTER TABLE plans DROP COLUMN IF EXISTS stripe_price_id;
