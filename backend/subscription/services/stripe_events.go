package services

import (
	"context"
	"database/sql"

	stripe "github.com/stripe/stripe-go/v82"
)

// ApplyStripeSubscriptionUpdate persists a Stripe subscription event to the DB
// and returns the current tier for the customer. Returns an empty tier (no
// error) when the customer has no row in our subscriptions table.
func ApplyStripeSubscriptionUpdate(ctx context.Context, db *sql.DB, sub stripe.Subscription) (string, error) {
	status := MapStripeStatus(string(sub.Status))

	var periodStart, periodEnd int64
	if sub.Items != nil && len(sub.Items.Data) > 0 {
		periodStart = sub.Items.Data[0].CurrentPeriodStart
		periodEnd = sub.Items.Data[0].CurrentPeriodEnd
	}

	var planID int
	if sub.Items != nil && len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
		_ = db.QueryRowContext(ctx,
			"SELECT plan_id FROM plans WHERE stripe_price_id = $1",
			sub.Items.Data[0].Price.ID,
		).Scan(&planID)
	}

	var query string
	var args []any
	if planID > 0 {
		query = `UPDATE subscriptions
		         SET stripe_subscription_id       = $1,
		             status                       = $2,
		             plan_id                      = $3,
		             current_period_start         = to_timestamp($4),
		             current_period_end           = to_timestamp($5),
		             cancel_at_period_end         = $6,
		             pending_plan_id              = NULL,
		             price_cents_at_subscription  = (SELECT price_cents FROM plans WHERE plan_id = $3),
		             updated_at                   = NOW()
		         WHERE stripe_customer_id = $7
		         RETURNING (SELECT tier FROM plans WHERE plan_id = $3)`
		args = []any{sub.ID, status, planID, periodStart, periodEnd, sub.CancelAtPeriodEnd, sub.Customer.ID}
	} else {
		query = `UPDATE subscriptions
		         SET stripe_subscription_id = $1,
		             status                 = $2,
		             current_period_start   = to_timestamp($3),
		             current_period_end     = to_timestamp($4),
		             cancel_at_period_end   = $5,
		             pending_plan_id        = NULL,
		             updated_at             = NOW()
		         WHERE stripe_customer_id = $6
		         RETURNING (SELECT tier FROM plans WHERE plan_id = subscriptions.plan_id)`
		args = []any{sub.ID, status, periodStart, periodEnd, sub.CancelAtPeriodEnd, sub.Customer.ID}
	}

	var tier string
	if err := db.QueryRowContext(ctx, query, args...).Scan(&tier); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return tier, nil
}

// MapStripeStatus converts a Stripe subscription status to the internal status.
func MapStripeStatus(stripeStatus string) string {
	switch stripeStatus {
	case "active", "trialing":
		return "active"
	case "canceled":
		return "cancelled"
	case "past_due", "unpaid":
		return "expired"
	default:
		return "active"
	}
}
