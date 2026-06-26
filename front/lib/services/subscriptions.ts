import { apiFetch, type ApiResult } from "@/lib/api";

export async function listPlans(): Promise<ApiResult> {
  return apiFetch("/api/plans");
}

export async function getMySubscription(): Promise<ApiResult> {
  return apiFetch("/api/subscriptions/me");
}

export async function listAdminSubscriptions(): Promise<ApiResult> {
  return apiFetch("/api/subscriptions/admin");
}

export async function assignPlan(userId: string, planId: number): Promise<ApiResult> {
  return apiFetch(
    `/api/subscriptions/admin/${encodeURIComponent(userId)}/plan`,
    {
      method: "POST",
      body: JSON.stringify({ plan_id: planId }),
    },
  );
}

export async function createPaymentIntent(planId: number): Promise<ApiResult> {
  return apiFetch("/api/subscriptions/payment-intent", {
    method: "POST",
    body: JSON.stringify({ plan_id: planId }),
  });
}

export async function cancelSubscription(): Promise<ApiResult> {
  return apiFetch("/api/subscriptions/cancel", { method: "POST" });
}

export async function getAdminStats(): Promise<ApiResult> {
  return apiFetch("/api/subscriptions/admin/stats");
}

export async function listAllPlans(): Promise<ApiResult> {
  return apiFetch("/api/plans?include_inactive=true");
}

export async function updatePlan(
  planId: number,
  data: {
    name: string;
    price_cents: number;
    billing_cycle: string;
    stripe_price_id: string;
    features: string;
  },
): Promise<ApiResult> {
  return apiFetch(`/api/plans/${planId}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}
