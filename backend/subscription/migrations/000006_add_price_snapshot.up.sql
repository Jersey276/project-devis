ALTER TABLE subscriptions ADD COLUMN price_cents_at_subscription INTEGER NOT NULL DEFAULT 0;

UPDATE subscriptions s
SET price_cents_at_subscription = p.price_cents
FROM plans p
WHERE p.plan_id = s.plan_id;
