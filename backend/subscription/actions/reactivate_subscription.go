package actions

import (
	"context"
	"database/sql"

	stripe "github.com/stripe/stripe-go/v82"
	stripesub "github.com/stripe/stripe-go/v82/subscription"
	subGrpc "project-devis-subscription/services/grpc"
)

func (s *Server) ReactivateSubscription(ctx context.Context, req *subGrpc.ReactivateSubscriptionRequest) (*subGrpc.GenericResponse, error) {
	if req.GetUserId() == "" {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	var stripeSubID sql.NullString
	var cancelAtPeriodEnd bool
	err := s.db.QueryRowContext(ctx,
		"SELECT stripe_subscription_id, cancel_at_period_end FROM subscriptions WHERE user_id = $1",
		req.GetUserId(),
	).Scan(&stripeSubID, &cancelAtPeriodEnd)
	if err == sql.ErrNoRows {
		return &subGrpc.GenericResponse{Success: false, Code: CodeNotFound}, nil
	}
	if err != nil {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInternalError}, nil
	}
	if !stripeSubID.Valid || stripeSubID.String == "" {
		return &subGrpc.GenericResponse{Success: false, Code: CodeNotFound}, nil
	}
	if !cancelAtPeriodEnd {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	_, err = stripesub.Update(stripeSubID.String, &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
	})
	if err != nil {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInternalError}, nil
	}

	_, err = s.db.ExecContext(ctx,
		"UPDATE subscriptions SET cancel_at_period_end = FALSE, updated_at = NOW() WHERE user_id = $1",
		req.GetUserId(),
	)
	if err != nil {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInternalError}, nil
	}

	return &subGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}
