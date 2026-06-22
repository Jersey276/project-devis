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

export async function listClients(): Promise<ApiResult> {
  return apiFetch("/api/users/clients");
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

export type { BackendClient, ClientType };
