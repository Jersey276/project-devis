import { apiFetch } from "@/lib/api";

export function listPlans() {
  return apiFetch("/api/plans");
}

export function getMySubscription() {
  return apiFetch("/api/subscriptions/me");
}

export function listAdminSubscriptions() {
  return apiFetch("/api/subscriptions/admin");
}

export function assignPlan(userId: string, planId: number) {
  return apiFetch(
    `/api/subscriptions/admin/${encodeURIComponent(userId)}/plan`,
    {
      method: "POST",
      body: JSON.stringify({ plan_id: planId }),
    },
  );
}

export function createPaymentIntent(planId: number) {
  return apiFetch("/api/subscriptions/payment-intent", {
    method: "POST",
    body: JSON.stringify({ plan_id: planId }),
  });
}

export function cancelSubscription() {
  return apiFetch("/api/subscriptions/cancel", { method: "POST" });
}

export function getAdminStats() {
  return apiFetch("/api/subscriptions/admin/stats");
}

export function listAllPlans() {
  return apiFetch("/api/plans?include_inactive=true");
}

export function updatePlan(
  planId: number,
  data: {
    name: string;
    price_cents: number;
    billing_cycle: string;
    stripe_price_id: string;
    features: string;
  },
) {
  return apiFetch(`/api/plans/${planId}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}
