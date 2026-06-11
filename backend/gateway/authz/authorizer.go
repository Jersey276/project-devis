package authz

import "context"

type Action string

type Resource string

const (
	ActionRead   Action = "read"
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionManage Action = "manage"
)

const (
	ResourceGeneral                   Resource = "general"
	ResourceAdminCountries            Resource = "admin.countries"
	ResourceAdminCountryGroup         Resource = "admin.country_groups"
	ResourceAdminTaxes                Resource = "admin.taxes"
	ResourceAdminSubscriptions        Resource = "admin.subscriptions"
	ResourceSubscriptionTemplates     Resource = "subscription.templates"
	ResourceSubscriptionSchedules     Resource = "subscription.schedules"
	ResourceSubscriptionEmailTracking Resource = "subscription.email_tracking"
)

type Subject struct {
	Role             string
	AccountStatus    string
	SubscriptionTier string
}

type Decision struct {
	Allowed bool
	Reason  string
}

type Authorizer interface {
	Can(ctx context.Context, subject Subject, action Action, resource Resource) (Decision, error)
}
