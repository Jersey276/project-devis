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

func unmarshalStripeData[T any](raw []byte) (T, *subGrpc.GenericResponse) {
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return v, &subGrpc.GenericResponse{Success: false, Code: CodeInternalError}
	}
	return v, nil
}

func (s *Server) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) (*subGrpc.GenericResponse, error) {
	sub, errResp := unmarshalStripeData[stripe.Subscription](event.Data.Raw)
	if errResp != nil {
		return errResp, nil
	}

	tier, err := services.ApplyStripeSubscriptionUpdate(ctx, s.db, sub)
	if err != nil {
		log.Printf("handleSubscriptionUpdated: no row for customer %s: %v", sub.Customer.ID, err)
		return &subGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
	}
	if tier != "" {
		s.syncAuthTier(ctx, sub.Customer.ID, tier)
	}
	return &subGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}

func (s *Server) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) (*subGrpc.GenericResponse, error) {
	sub, errResp := unmarshalStripeData[stripe.Subscription](event.Data.Raw)
	if errResp != nil {
		return errResp, nil
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
	invoice, errResp := unmarshalStripeData[stripe.Invoice](event.Data.Raw)
	if errResp != nil {
		return errResp, nil
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

