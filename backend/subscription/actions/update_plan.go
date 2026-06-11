package actions

import (
	"context"
	"database/sql"
	"encoding/json"

	subscriptionGrpc "project-devis-subscription/services/grpc"
)

func (s *Server) UpdatePlan(ctx context.Context, req *subscriptionGrpc.UpdatePlanRequest) (*subscriptionGrpc.GetPlanResponse, error) {
	if req.GetPlanId() == 0 {
		return &subscriptionGrpc.GetPlanResponse{Success: false, Code: CodeInvalidInput}, nil
	}
	if req.GetName() == "" {
		return &subscriptionGrpc.GetPlanResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	// Validate features is valid JSON (empty string → default to {})
	features := req.GetFeatures()
	if features == "" {
		features = "{}"
	}
	if !json.Valid([]byte(features)) {
		return &subscriptionGrpc.GetPlanResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	var stripePriceID sql.NullString
	if v := req.GetStripePriceId(); v != "" {
		stripePriceID = sql.NullString{String: v, Valid: true}
	}

	_, err := s.db.ExecContext(ctx,
		`UPDATE plans
		 SET name            = $1,
		     price_cents     = $2,
		     billing_cycle   = $3,
		     stripe_price_id = $4,
		     features        = $5::jsonb
		 WHERE plan_id = $6`,
		req.GetName(), req.GetPriceCents(), req.GetBillingCycle(),
		stripePriceID, features, req.GetPlanId(),
	)
	if err != nil {
		return &subscriptionGrpc.GetPlanResponse{Success: false, Code: CodeInternalError}, nil
	}

	p := &subscriptionGrpc.Plan{}
	var spid sql.NullString
	err = s.db.QueryRowContext(ctx,
		`SELECT plan_id, name, tier, price_cents, billing_cycle, features::text, active, COALESCE(stripe_price_id, '')
		 FROM plans WHERE plan_id = $1`,
		req.GetPlanId(),
	).Scan(&p.PlanId, &p.Name, &p.Tier, &p.PriceCents, &p.BillingCycle, &p.Features, &p.Active, &spid)
	if code, isErr := queryErrCode(err); isErr {
		return &subscriptionGrpc.GetPlanResponse{Success: false, Code: code}, nil
	}
	p.StripePriceId = spid.String

	return &subscriptionGrpc.GetPlanResponse{Success: true, Code: CodeSuccess, Plan: p}, nil
}
