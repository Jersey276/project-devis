package actions

import (
	"context"
	"database/sql"

	stripe "github.com/stripe/stripe-go/v82"
	stripesub "github.com/stripe/stripe-go/v82/subscription"
	subGrpc "project-devis-subscription/services/grpc"
)

func (s *Server) CancelSubscription(ctx context.Context, req *subGrpc.CancelSubscriptionRequest) (*subGrpc.GenericResponse, error) {
	if req.GetUserId() == "" {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	var stripeSubID sql.NullString
	err := s.db.QueryRowContext(ctx,
		"SELECT stripe_subscription_id FROM subscriptions WHERE user_id = $1",
		req.GetUserId(),
	).Scan(&stripeSubID)
	if err == sql.ErrNoRows || !stripeSubID.Valid || stripeSubID.String == "" {
		return &subGrpc.GenericResponse{Success: false, Code: CodeNotFound}, nil
	}
	if err != nil {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInternalError}, nil
	}

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}
	if _, err := stripesub.Update(stripeSubID.String, params); err != nil {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInternalError}, nil
	}

	_, err = s.db.ExecContext(ctx,
		"UPDATE subscriptions SET cancel_at_period_end = TRUE, updated_at = NOW() WHERE user_id = $1",
		req.GetUserId(),
	)
	if err != nil {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInternalError}, nil
	}

	return &subGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}
