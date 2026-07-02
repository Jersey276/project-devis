package actions

import (
	"context"

	subGrpc "project-devis-subscription/services/grpc"
)

func (s *Server) GetAdminStats(ctx context.Context, _ *subGrpc.GetAdminStatsRequest) (*subGrpc.AdminStatsResponse, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.tier, COUNT(s.subscription_id)
		FROM plans p
		LEFT JOIN subscriptions s ON s.plan_id = p.plan_id AND s.status = 'active'
		GROUP BY p.tier
	`)
	if err != nil {
		return &subGrpc.AdminStatsResponse{Success: false, Code: CodeInternalError}, nil
	}
	defer rows.Close()

	var distribution []*subGrpc.PlanDistributionEntry
	for rows.Next() {
		entry := &subGrpc.PlanDistributionEntry{}
		if err := rows.Scan(&entry.Tier, &entry.Count); err != nil {
			return &subGrpc.AdminStatsResponse{Success: false, Code: CodeInternalError}, nil
		}
		distribution = append(distribution, entry)
	}

	var totalActive int32
	var totalRevenue int64
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(s.price_cents_at_subscription), 0)
		FROM subscriptions s
		WHERE s.status IN ('active', 'cancelled')
	`).Scan(&totalActive, &totalRevenue)
	if err != nil {
		return &subGrpc.AdminStatsResponse{Success: false, Code: CodeInternalError}, nil
	}

	monthRows, err := s.db.QueryContext(ctx, `
		SELECT date_trunc('month', s.current_period_start)::date::text AS month,
		       SUM(s.price_cents_at_subscription) AS revenue_cents
		FROM subscriptions s
		WHERE s.status = 'active'
		  AND s.current_period_start >= date_trunc('month', NOW() - INTERVAL '11 months')
		GROUP BY 1
		ORDER BY 1
	`)
	if err != nil {
		return &subGrpc.AdminStatsResponse{Success: false, Code: CodeInternalError}, nil
	}
	defer monthRows.Close()

	var monthly []*subGrpc.MonthlyRevenueEntry
	for monthRows.Next() {
		entry := &subGrpc.MonthlyRevenueEntry{}
		if err := monthRows.Scan(&entry.Month, &entry.RevenueCents); err != nil {
			return &subGrpc.AdminStatsResponse{Success: false, Code: CodeInternalError}, nil
		}
		monthly = append(monthly, entry)
	}

	return &subGrpc.AdminStatsResponse{
		Success:                  true,
		Code:                     CodeSuccess,
		TotalActiveSubscriptions: totalActive,
		TotalRevenueCents:        totalRevenue,
		PlanDistribution:         distribution,
		MonthlyRevenue:           monthly,
	}, nil
}
