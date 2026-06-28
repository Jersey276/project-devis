package actions

import (
	"context"
	"database/sql"
	"fmt"

	stripe "github.com/stripe/stripe-go/v82"
	stripesub "github.com/stripe/stripe-go/v82/subscription"
	subGrpc "project-devis-subscription/services/grpc"
)

func (s *Server) ChangePlan(ctx context.Context, req *subGrpc.ChangePlanRequest) (*subGrpc.GenericResponse, error) {
	if req.GetUserId() == "" || req.GetPlanId() == 0 {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}
	if stripe.Key == "" {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInternalError}, fmt.Errorf("stripe key not configured")
	}

	var stripeSubID sql.NullString
	var currentPlanID int32
	var cancelAtPeriodEnd bool
	err := s.db.QueryRowContext(ctx,
		`SELECT stripe_subscription_id, plan_id, cancel_at_period_end
		 FROM subscriptions WHERE user_id = $1`,
		req.GetUserId(),
	).Scan(&stripeSubID, &currentPlanID, &cancelAtPeriodEnd)
	if err == sql.ErrNoRows {
		return &subGrpc.GenericResponse{Success: false, Code: CodeNotFound}, nil
	}
	if err != nil {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInternalError}, nil
	}
	if !stripeSubID.Valid || stripeSubID.String == "" {
		return &subGrpc.GenericResponse{Success: false, Code: CodeNotFound}, nil
	}
	if cancelAtPeriodEnd {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}
	if currentPlanID == req.GetPlanId() {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	var newStripePriceID string
	err = s.db.QueryRowContext(ctx,
		"SELECT stripe_price_id FROM plans WHERE plan_id = $1 AND active = TRUE",
		req.GetPlanId(),
	).Scan(&newStripePriceID)
	if err == sql.ErrNoRows || newStripePriceID == "" {
		return &subGrpc.GenericResponse{Success: false, Code: CodeNotFound}, nil
	}
	if err != nil {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInternalError}, nil
	}

	stripeSub, err := stripesub.Get(stripeSubID.String, nil)
	if err != nil || len(stripeSub.Items.Data) == 0 {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInternalError}, nil
	}
	itemID := stripeSub.Items.Data[0].ID

	_, err = stripesub.Update(stripeSubID.String, &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{{
			ID:    stripe.String(itemID),
			Price: stripe.String(newStripePriceID),
		}},
		ProrationBehavior: stripe.String("none"),
	})
	if err != nil {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInternalError}, nil
	}

	_, err = s.db.ExecContext(ctx,
		"UPDATE subscriptions SET pending_plan_id = $1, updated_at = NOW() WHERE user_id = $2",
		req.GetPlanId(), req.GetUserId(),
	)
	if err != nil {
		return &subGrpc.GenericResponse{Success: false, Code: CodeInternalError}, nil
	}

	return &subGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}
