import { apiFetch, type ApiResult } from "@/lib/api";
import type { BackendClient, ClientType } from "@/types/backend";

// Full-replace shape: all fields are required strings on the wire.
// Empty string clears the column server-side (see client.Update action).
export type ClientPayload = {
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  company: string;
  siren: string;
  siret: string;
  vat: string;
  client_type: ClientType;
};

export async function listClients(queryString?: string, signal?: AbortSignal): Promise<ApiResult> {
  const url = queryString ? `/api/users/clients?${queryString}` : "/api/users/clients";
  return apiFetch(url, { signal });
}

export async function createClient(payload: ClientPayload): Promise<ApiResult> {
  return apiFetch("/api/users/clients", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function getClient(clientId: string): Promise<ApiResult> {
  return apiFetch(`/api/users/clients/${encodeURIComponent(clientId)}`);
}

export async function updateClient(
  clientId: string,
  payload: ClientPayload,
): Promise<ApiResult> {
  return apiFetch(`/api/users/clients/${encodeURIComponent(clientId)}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export async function archiveClient(clientId: string): Promise<ApiResult> {
  return apiFetch(`/api/users/clients/${encodeURIComponent(clientId)}`, {
    method: "DELETE",
  });
}

export async function sendClientInvitation(clientId: string): Promise<ApiResult> {
  return apiFetch("/api/auth/invite/client", {
    method: "POST",
    body: JSON.stringify({ client_id: clientId }),
  });
}

export async function getMyClientProfiles(): Promise<ApiResult> {
  return apiFetch("/api/users/clients/me");
}

export async function updateMyClientProfile(clientId: string, payload: ClientPayload): Promise<ApiResult> {
  return apiFetch(`/api/users/clients/me?client_id=${encodeURIComponent(clientId)}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export async function getMyClientAddresses(clientId: string): Promise<ApiResult> {
  return apiFetch(`/api/users/clients/me/addresses?client_id=${encodeURIComponent(clientId)}`);
}

export type { BackendClient, ClientType };
