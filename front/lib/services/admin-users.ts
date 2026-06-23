import { apiFetch } from "@/lib/api";

export function listAdminUsers(params?: string) {
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

export function updateAdminUser(
  userId: string,
  payload: UpdateAdminUserPayload,
) {
  return apiFetch(`/api/users/admin/accounts/${encodeURIComponent(userId)}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export function suspendAdminUser(userId: string) {
  return apiFetch(
    `/api/users/admin/accounts/${encodeURIComponent(userId)}/suspend`,
    {
      method: "POST",
    },
  );
}
