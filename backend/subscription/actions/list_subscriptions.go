package actions

import (
	"context"
	"database/sql"

	subscriptionGrpc "project-devis-subscription/services/grpc"
)

func (s *Server) ListSubscriptions(ctx context.Context, req *subscriptionGrpc.ListSubscriptionsRequest) (*subscriptionGrpc.ListSubscriptionsResponse, error) {
	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 20
	}
	page := req.GetPage()
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	var total int32
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM subscriptions").Scan(&total); err != nil {
		return &subscriptionGrpc.ListSubscriptionsResponse{Success: false, Code: CodeInternalError}, nil
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT s.subscription_id, s.user_id, s.plan_id, p.tier, s.status,
		        s.current_period_start::text, s.current_period_end::text,
		        s.created_at::text, s.updated_at::text
		 FROM subscriptions s
		 JOIN plans p ON p.plan_id = s.plan_id
		 ORDER BY s.updated_at DESC
		 LIMIT $1 OFFSET $2`,
		pageSize, offset,
	)
	if err != nil {
		return &subscriptionGrpc.ListSubscriptionsResponse{Success: false, Code: CodeInternalError}, nil
	}
	defer rows.Close()

	var subs []*subscriptionGrpc.Subscription
	for rows.Next() {
		sub := &subscriptionGrpc.Subscription{}
		var periodEnd sql.NullString
		if err := rows.Scan(
			&sub.SubscriptionId, &sub.UserId, &sub.PlanId, &sub.Tier, &sub.Status,
			&sub.CurrentPeriodStart, &periodEnd,
			&sub.CreatedAt, &sub.UpdatedAt,
		); err != nil {
			return &subscriptionGrpc.ListSubscriptionsResponse{Success: false, Code: CodeInternalError}, nil
		}
		if periodEnd.Valid {
			sub.CurrentPeriodEnd = periodEnd.String
		}
		subs = append(subs, sub)
	}

	return &subscriptionGrpc.ListSubscriptionsResponse{
		Success:       true,
		Code:          CodeSuccess,
		Subscriptions: subs,
		Total:         total,
	}, nil
}
