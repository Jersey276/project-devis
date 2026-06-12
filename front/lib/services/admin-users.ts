import { apiFetch } from "@/lib/api";

export function listAdminUsers() {
  return apiFetch("/api/users/admin/accounts");
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
