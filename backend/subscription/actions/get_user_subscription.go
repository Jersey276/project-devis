package actions

import (
	"context"
	"database/sql"

	subscriptionGrpc "project-devis-subscription/services/grpc"
)

func (s *Server) GetUserSubscription(ctx context.Context, req *subscriptionGrpc.GetUserSubscriptionRequest) (*subscriptionGrpc.GetUserSubscriptionResponse, error) {
	sub := &subscriptionGrpc.Subscription{}
	var periodEnd sql.NullString

	err := s.db.QueryRowContext(ctx,
		`SELECT s.subscription_id, s.user_id, s.plan_id, p.tier, s.status,
		        s.current_period_start::text, s.current_period_end::text,
		        s.created_at::text, s.updated_at::text
		 FROM subscriptions s
		 JOIN plans p ON p.plan_id = s.plan_id
		 WHERE s.user_id = $1`,
		req.GetUserId(),
	).Scan(
		&sub.SubscriptionId, &sub.UserId, &sub.PlanId, &sub.Tier, &sub.Status,
		&sub.CurrentPeriodStart, &periodEnd,
		&sub.CreatedAt, &sub.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Return a synthetic free subscription for users without a record.
		return &subscriptionGrpc.GetUserSubscriptionResponse{
			Success: true,
			Code:    CodeSuccess,
			Subscription: &subscriptionGrpc.Subscription{
				UserId: req.GetUserId(),
				Tier:   "free",
				Status: "active",
			},
		}, nil
	}
	if err != nil {
		return &subscriptionGrpc.GetUserSubscriptionResponse{Success: false, Code: CodeInternalError}, nil
	}

	if periodEnd.Valid {
		sub.CurrentPeriodEnd = periodEnd.String
	}

	return &subscriptionGrpc.GetUserSubscriptionResponse{Success: true, Code: CodeSuccess, Subscription: sub}, nil
}
