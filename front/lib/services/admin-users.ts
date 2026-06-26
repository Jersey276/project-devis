import { apiFetch, type ApiResult } from "@/lib/api";

export async function getAdminUser(userId: string): Promise<ApiResult> {
  return apiFetch(`/api/users/admin/accounts/${encodeURIComponent(userId)}`);
}

export async function listAdminUsers(params?: string): Promise<ApiResult> {
  const url = params
    ? `/api/users/admin/accounts?${params}`
    : "/api/users/admin/accounts";
  return apiFetch(url);
}

type UpdateAdminUserPayload = {
  first_name: string;
  last_name: string;
  email: string;
  role: "user" | "admin";
  plan: string;
  phone?: string;
  company?: string;
  siren?: string;
  vat?: string;
};

export async function updateAdminUser(
  userId: string,
  payload: UpdateAdminUserPayload,
): Promise<ApiResult> {
  return apiFetch(`/api/users/admin/accounts/${encodeURIComponent(userId)}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export async function suspendAdminUser(userId: string): Promise<ApiResult> {
  return apiFetch(
    `/api/users/admin/accounts/${encodeURIComponent(userId)}/suspend`,
    {
      method: "POST",
    },
  );
}

export async function listAdminUserQuotes(userId: string, params?: string): Promise<ApiResult> {
  const base = `/api/users/admin/accounts/${encodeURIComponent(userId)}/quotes`;
  return apiFetch(params ? `${base}?${params}` : base);
}

export async function listAdminUserSchedules(userId: string, params?: string): Promise<ApiResult> {
  const base = `/api/users/admin/accounts/${encodeURIComponent(userId)}/schedules`;
  return apiFetch(params ? `${base}?${params}` : base);
}

export async function listAdminUserInvoices(userId: string, params?: string): Promise<ApiResult> {
  const base = `/api/users/admin/accounts/${encodeURIComponent(userId)}/invoices`;
  return apiFetch(params ? `${base}?${params}` : base);
}
