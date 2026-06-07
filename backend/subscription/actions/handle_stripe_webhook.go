package actions

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
	authpb "project-devis-subscription/services/grpc/auth"
	"project-devis-subscription/services"
	subGrpc "project-devis-subscription/services/grpc"
)

func (s *Server) HandleStripeWebhook(ctx context.Context, req *subGrpc.HandleStripeWebhookRequest) (*subGrpc.GenericResponse, error) {
	event, err := webhook.ConstructEventWithOptions(
		req.GetPayload(),
		req.GetStripeSignature(),
		s.webhookSecret,
		webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true},
	)
	if err != nil {
		log.Printf("webhook.ConstructEvent failed: %v", err)
		return &subGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}
	return s.ProcessWebhookEvent(ctx, event)
}

func (s *Server) ProcessWebhookEvent(ctx context.Context, event stripe.Event) (*subGrpc.GenericResponse, error) {
	eventID := uuid.New().String()
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO subscription_events (event_id, user_id, stripe_event_id, event_type, payload)
		 VALUES ($1, '', $2, $3, $4)
		 ON CONFLICT (stripe_event_id) DO NOTHING`,
		eventID, event.ID, event.Type, string(event.Data.Raw),
	)
	if err != nil {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInternalError}, nil
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return &subGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
	}

	switch event.Type {
	case "customer.subscription.updated", "customer.subscription.created":
		return s.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(ctx, event)
	case "invoice.payment_failed":
		return s.handleInvoicePaymentFailed(ctx, event)
	}

	return &subGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

func (s *Server) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) (*subGrpc.GenericResponse, error) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInternalError}, nil
	}

	status := mapStripeStatus(string(sub.Status))

	// In stripe-go v82, period dates are on subscription items, not the subscription itself.
	var periodStart, periodEnd int64
	if sub.Items != nil && len(sub.Items.Data) > 0 {
		periodStart = sub.Items.Data[0].CurrentPeriodStart
		periodEnd = sub.Items.Data[0].CurrentPeriodEnd
	}

	// Resolve plan_id from the Stripe price ID on the first subscription item.
	var planID int
	if sub.Items != nil && len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
		_ = s.db.QueryRowContext(ctx,
			"SELECT plan_id FROM plans WHERE stripe_price_id = $1",
			sub.Items.Data[0].Price.ID,
		).Scan(&planID)
	}

	var currentTier string
	var query string
	var args []any
	if planID > 0 {
		query = `UPDATE subscriptions
		         SET stripe_subscription_id = $1,
		             status                 = $2,
		             plan_id                = $3,
		             current_period_start   = to_timestamp($4),
		             current_period_end     = to_timestamp($5),
		             cancel_at_period_end   = $6,
		             updated_at             = NOW()
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
		             updated_at             = NOW()
		         WHERE stripe_customer_id = $6
		         RETURNING (SELECT tier FROM plans WHERE plan_id = subscriptions.plan_id)`
		args = []any{sub.ID, status, periodStart, periodEnd, sub.CancelAtPeriodEnd, sub.Customer.ID}
	}

	err := s.db.QueryRowContext(ctx, query, args...).Scan(&currentTier)
	if err != nil {
		log.Printf("handleSubscriptionUpdated: no row for customer %s: %v", sub.Customer.ID, err)
		return &subGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
	}

	s.syncAuthTier(ctx, sub.Customer.ID, currentTier)
	return &subGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

func (s *Server) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) (*subGrpc.GenericResponse, error) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInternalError}, nil
	}

	var userID string
	err := s.db.QueryRowContext(ctx,
		`UPDATE subscriptions SET status = 'cancelled', updated_at = NOW()
		 WHERE stripe_customer_id = $1
		 RETURNING user_id`,
		sub.Customer.ID,
	).Scan(&userID)
	if err != nil {
		return &subGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
	}

	s.callAuthUpdateTier(ctx, userID, "free")
	return &subGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

func (s *Server) handleInvoicePaymentFailed(ctx context.Context, event stripe.Event) (*subGrpc.GenericResponse, error) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInternalError}, nil
	}

	if invoice.Customer == nil {
		return &subGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
	}

	_, err := s.db.ExecContext(ctx,
		"UPDATE subscriptions SET status = 'expired', updated_at = NOW() WHERE stripe_customer_id = $1",
		invoice.Customer.ID,
	)
	if err != nil {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInternalError}, nil
	}

	return &subGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

func (s *Server) syncAuthTier(ctx context.Context, stripeCustomerID, tier string) {
	var userID string
	err := s.db.QueryRowContext(ctx,
		"SELECT user_id FROM subscriptions WHERE stripe_customer_id = $1",
		stripeCustomerID,
	).Scan(&userID)
	if err != nil {
		return
	}
	s.callAuthUpdateTier(ctx, userID, tier)
}

func (s *Server) callAuthUpdateTier(ctx context.Context, userID, tier string) {
	client, err := services.GetAuthServiceClient()
	if err != nil {
		log.Printf("callAuthUpdateTier: auth client unavailable: %v", err)
		return
	}
	if _, err := client.UpdateSubscriptionTier(ctx, &authpb.UpdateSubscriptionTierRequest{
		UserId: userID,
		Tier:   tier,
	}); err != nil {
		log.Printf("callAuthUpdateTier: failed for user %s tier %s: %v", userID, tier, err)
	}
}

func mapStripeStatus(stripeStatus string) string {
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
