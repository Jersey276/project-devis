package actions

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/customer"
	stripesub "github.com/stripe/stripe-go/v82/subscription"
	subGrpc "project-devis-subscription/services/grpc"
)

func (s *Server) CreatePaymentIntent(ctx context.Context, req *subGrpc.CreatePaymentIntentRequest) (*subGrpc.CreatePaymentIntentResponse, error) {
	if req.GetUserId() == "" || req.GetPlanId() == 0 {
		return &subGrpc.CreatePaymentIntentResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	var stripePriceID sql.NullString
	err := s.db.QueryRowContext(ctx,
		"SELECT stripe_price_id FROM plans WHERE plan_id = $1 AND active = TRUE",
		req.GetPlanId(),
	).Scan(&stripePriceID)
	if code, isErr := queryErrCode(err); isErr {
		return &subGrpc.CreatePaymentIntentResponse{Success: false, Code: code}, nil
	}
	if !stripePriceID.Valid || stripePriceID.String == "" {
		return &subGrpc.CreatePaymentIntentResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	var stripeCustomerID sql.NullString
	rowErr := s.db.QueryRowContext(ctx,
		"SELECT stripe_customer_id FROM subscriptions WHERE user_id = $1",
		req.GetUserId(),
	).Scan(&stripeCustomerID)

	customerID := ""
	if rowErr == nil && stripeCustomerID.Valid && stripeCustomerID.String != "" {
		customerID = stripeCustomerID.String
	} else {
		email := req.GetEmail()
		params := &stripe.CustomerParams{Email: &email}
		c, stripeErr := customer.New(params)
		if stripeErr != nil {
			return &subGrpc.CreatePaymentIntentResponse{Success: false, Code: CodeInternalError}, nil
		}
		customerID = c.ID

		subID := uuid.New().String()
		if _, dbErr := s.db.ExecContext(ctx,
			`INSERT INTO subscriptions (subscription_id, user_id, plan_id, status, stripe_customer_id)
			 VALUES ($1, $2, $3, 'active', $4)
			 ON CONFLICT (user_id) DO UPDATE SET stripe_customer_id = EXCLUDED.stripe_customer_id, updated_at = NOW()`,
			subID, req.GetUserId(), req.GetPlanId(), customerID,
		); dbErr != nil {
			return &subGrpc.CreatePaymentIntentResponse{Success: false, Code: CodeInternalError}, nil
		}
	}

	subParams := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{Price: stripe.String(stripePriceID.String)},
		},
		PaymentBehavior: stripe.String("default_incomplete"),
	}
	subParams.AddExpand("latest_invoice.confirmation_secret")

	stripeSub, stripeErr := stripesub.New(subParams)
	if stripeErr != nil {
		return &subGrpc.CreatePaymentIntentResponse{Success: false, Code: CodeInternalError}, nil
	}

	if stripeSub.LatestInvoice == nil || stripeSub.LatestInvoice.ConfirmationSecret == nil {
		return &subGrpc.CreatePaymentIntentResponse{Success: false, Code: CodeInternalError}, nil
	}

	return &subGrpc.CreatePaymentIntentResponse{
		Success:              true,
		Code:                 CodeSuccess,
		ClientSecret:         stripeSub.LatestInvoice.ConfirmationSecret.ClientSecret,
		StripeSubscriptionId: stripeSub.ID,
	}, nil
}
