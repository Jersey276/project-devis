import { apiFetch, type ApiResult } from "@/lib/api";

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
