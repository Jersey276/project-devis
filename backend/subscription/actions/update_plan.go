package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	stripe "github.com/stripe/stripe-go/v82"
	stripeprice "github.com/stripe/stripe-go/v82/price"
	stripeproduct "github.com/stripe/stripe-go/v82/product"
	subscriptionGrpc "project-devis-subscription/services/grpc"
)

func (s *Server) UpdatePlan(ctx context.Context, req *subscriptionGrpc.UpdatePlanRequest) (*subscriptionGrpc.GetPlanResponse, error) {
	if req.GetPlanId() == 0 {
		return &subscriptionGrpc.GetPlanResponse{Success: false, Code: CodeInvalidInput}, nil
	}
	if req.GetName() == "" {
		return &subscriptionGrpc.GetPlanResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	features := req.GetFeatures()
	if features == "" {
		features = "{}"
	}
	if !json.Valid([]byte(features)) {
		return &subscriptionGrpc.GetPlanResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	// Read current plan state before applying changes
	var cur struct {
		name          string
		priceCents    int32
		billingCycle  string
		stripePriceID sql.NullString
		stripeProduct sql.NullString
	}
	err := s.db.QueryRowContext(ctx,
		`SELECT name, price_cents, billing_cycle, stripe_price_id, stripe_product_id FROM plans WHERE plan_id = $1`,
		req.GetPlanId(),
	).Scan(&cur.name, &cur.priceCents, &cur.billingCycle, &cur.stripePriceID, &cur.stripeProduct)
	if code, isErr := queryErrCode(err); isErr {
		return &subscriptionGrpc.GetPlanResponse{Success: false, Code: code}, nil
	}

	newPriceID := cur.stripePriceID.String
	newProductID := cur.stripeProduct.String

	pricingChanged := req.GetPriceCents() != cur.priceCents || req.GetBillingCycle() != cur.billingCycle
	nameChanged := req.GetName() != cur.name
	isPaidPlan := req.GetBillingCycle() != "none"

	if stripe.Key != "" && isPaidPlan && (pricingChanged || nameChanged || !cur.stripeProduct.Valid) {
		if !cur.stripeProduct.Valid || cur.stripeProduct.String == "" {
			// First time: create Product + Price from scratch
			prod, stripeErr := stripeproduct.New(&stripe.ProductParams{
				Name: stripe.String(req.GetName()),
			})
			if stripeErr != nil {
				log.Printf("UpdatePlan: failed to create Stripe product: %v", stripeErr)
				return &subscriptionGrpc.GetPlanResponse{Success: false, Code: CodeInternalError}, nil
			}
			newProductID = prod.ID

			interval := billingCycleToStripeInterval(req.GetBillingCycle())
			pr, stripeErr := stripeprice.New(&stripe.PriceParams{
				Product:    stripe.String(newProductID),
				UnitAmount: stripe.Int64(int64(req.GetPriceCents())),
				Currency:   stripe.String("eur"),
				Recurring: &stripe.PriceRecurringParams{
					Interval: stripe.String(interval),
				},
			})
			if stripeErr != nil {
				log.Printf("UpdatePlan: failed to create Stripe price: %v", stripeErr)
				return &subscriptionGrpc.GetPlanResponse{Success: false, Code: CodeInternalError}, nil
			}
			newPriceID = pr.ID
		} else {
			// Existing product
			if nameChanged {
				if _, stripeErr := stripeproduct.Update(cur.stripeProduct.String, &stripe.ProductParams{
					Name: stripe.String(req.GetName()),
				}); stripeErr != nil {
					log.Printf("UpdatePlan: failed to update Stripe product name: %v", stripeErr)
					return &subscriptionGrpc.GetPlanResponse{Success: false, Code: CodeInternalError}, nil
				}
			}

			if pricingChanged {
				interval := billingCycleToStripeInterval(req.GetBillingCycle())
				pr, stripeErr := stripeprice.New(&stripe.PriceParams{
					Product:    stripe.String(cur.stripeProduct.String),
					UnitAmount: stripe.Int64(int64(req.GetPriceCents())),
					Currency:   stripe.String("eur"),
					Recurring: &stripe.PriceRecurringParams{
						Interval: stripe.String(interval),
					},
				})
				if stripeErr != nil {
					log.Printf("UpdatePlan: failed to create new Stripe price: %v", stripeErr)
					return &subscriptionGrpc.GetPlanResponse{Success: false, Code: CodeInternalError}, nil
				}
				newPriceID = pr.ID

				// Archive old price — non-blocking if it fails
				if cur.stripePriceID.Valid && cur.stripePriceID.String != "" {
					active := false
					if _, archErr := stripeprice.Update(cur.stripePriceID.String, &stripe.PriceParams{
						Active: &active,
					}); archErr != nil {
						log.Printf("UpdatePlan: warning — failed to archive old Stripe price %s: %v", cur.stripePriceID.String, archErr)
					}
				}
			}
		}
	}

	var stripePriceID sql.NullString
	if newPriceID != "" {
		stripePriceID = sql.NullString{String: newPriceID, Valid: true}
	}
	var stripeProductID sql.NullString
	if newProductID != "" {
		stripeProductID = sql.NullString{String: newProductID, Valid: true}
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE plans
		 SET name              = $1,
		     price_cents       = $2,
		     billing_cycle     = $3,
		     stripe_price_id   = $4,
		     stripe_product_id = $5,
		     features          = $6::jsonb
		 WHERE plan_id = $7`,
		req.GetName(), req.GetPriceCents(), req.GetBillingCycle(),
		stripePriceID, stripeProductID, features, req.GetPlanId(),
	)
	if err != nil {
		return &subscriptionGrpc.GetPlanResponse{Success: false, Code: CodeInternalError}, nil
	}

	p := &subscriptionGrpc.Plan{}
	err = s.db.QueryRowContext(ctx,
		`SELECT plan_id, name, tier, price_cents, billing_cycle, features::text, active,
		        COALESCE(stripe_price_id, ''), COALESCE(stripe_product_id, '')
		 FROM plans WHERE plan_id = $1`,
		req.GetPlanId(),
	).Scan(&p.PlanId, &p.Name, &p.Tier, &p.PriceCents, &p.BillingCycle, &p.Features, &p.Active, &p.StripePriceId, &p.StripeProductId)
	if code, isErr := queryErrCode(err); isErr {
		return &subscriptionGrpc.GetPlanResponse{Success: false, Code: code}, nil
	}

	return &subscriptionGrpc.GetPlanResponse{Success: true, Code: CodeSuccess, Plan: p}, nil
}

func billingCycleToStripeInterval(cycle string) string {
	switch cycle {
	case "monthly":
		return "month"
	case "yearly":
		return "year"
	default:
		return "month"
	}
}
