package actions

import (
	"context"
	"github.com/google/uuid"
	subscriptionGrpc "project-devis-subscription/services/grpc"
)

func (s *Server) AssignPlan(ctx context.Context, req *subscriptionGrpc.AssignPlanRequest) (*subscriptionGrpc.AssignPlanResponse, error) {
	if req.GetUserId() == "" || req.GetPlanId() == 0 {
		return &subscriptionGrpc.AssignPlanResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	var newTier string
	var priceCents int32
	err := s.db.QueryRowContext(ctx, "SELECT tier, price_cents FROM plans WHERE plan_id = $1 AND active = TRUE", req.GetPlanId()).Scan(&newTier, &priceCents)
	if code, isErr := queryErrCode(err); isErr {
		return &subscriptionGrpc.AssignPlanResponse{Success: false, Code: code}, nil
	}

	subscriptionID := uuid.New().String()
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO subscriptions (subscription_id, user_id, plan_id, status, current_period_start, price_cents_at_subscription)
		 VALUES ($1, $2, $3, 'active', NOW(), $4)
		 ON CONFLICT (user_id) DO UPDATE
		 SET plan_id = EXCLUDED.plan_id,
		     status  = 'active',
		     price_cents_at_subscription = EXCLUDED.price_cents_at_subscription,
		     updated_at = NOW()`,
		subscriptionID, req.GetUserId(), req.GetPlanId(), priceCents,
	)
	if err != nil {
		return &subscriptionGrpc.AssignPlanResponse{Success: false, Code: CodeInternalError}, nil
	}

	return &subscriptionGrpc.AssignPlanResponse{Success: true, Code: CodeSuccess, NewTier: newTier}, nil
}
