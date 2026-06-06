package actions

import (
	"context"
	"database/sql"

	subscriptionGrpc "project-devis-subscription/services/grpc"
)

func (s *Server) GetPlan(ctx context.Context, req *subscriptionGrpc.GetPlanRequest) (*subscriptionGrpc.GetPlanResponse, error) {
	p := &subscriptionGrpc.Plan{}
	var spid sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT plan_id, name, tier, price_cents, billing_cycle, features::text, active, COALESCE(stripe_price_id, '')
		 FROM plans WHERE plan_id = $1`,
		req.GetPlanId(),
	).Scan(&p.PlanId, &p.Name, &p.Tier, &p.PriceCents, &p.BillingCycle, &p.Features, &p.Active, &spid)

	if err == sql.ErrNoRows {
		return &subscriptionGrpc.GetPlanResponse{Success: false, Code: CodeNotFound}, nil
	}
	if err != nil {
		return &subscriptionGrpc.GetPlanResponse{Success: false, Code: CodeInternalError}, nil
	}
	p.StripePriceId = spid.String

	return &subscriptionGrpc.GetPlanResponse{Success: true, Code: CodeSuccess, Plan: p}, nil
}
