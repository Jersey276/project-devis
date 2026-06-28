package actions

import (
	"context"

	subscriptionGrpc "project-devis-subscription/services/grpc"
)

func (s *Server) ListPlans(ctx context.Context, req *subscriptionGrpc.ListPlansRequest) (*subscriptionGrpc.ListPlansResponse, error) {
	query := `SELECT plan_id, name, tier, price_cents, billing_cycle, features::text, active, COALESCE(stripe_price_id, ''), COALESCE(stripe_product_id, '')
	          FROM plans`
	if !req.GetIncludeInactive() {
		query += " WHERE active = TRUE"
	}
	query += " ORDER BY price_cents ASC"

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return &subscriptionGrpc.ListPlansResponse{Success: false, Code: CodeInternalError}, nil
	}
	defer rows.Close()

	var plans []*subscriptionGrpc.Plan
	for rows.Next() {
		p := &subscriptionGrpc.Plan{}
		if err := rows.Scan(&p.PlanId, &p.Name, &p.Tier, &p.PriceCents, &p.BillingCycle, &p.Features, &p.Active, &p.StripePriceId, &p.StripeProductId); err != nil {
			return &subscriptionGrpc.ListPlansResponse{Success: false, Code: CodeInternalError}, nil
		}
		plans = append(plans, p)
	}

	return &subscriptionGrpc.ListPlansResponse{Success: true, Code: CodeSuccess, Plans: plans}, nil
}
