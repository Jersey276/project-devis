ALTER TABLE subscriptions ADD COLUMN pending_plan_id INTEGER REFERENCES plans(plan_id);
