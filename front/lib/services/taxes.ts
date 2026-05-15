import { apiFetch, type ApiResult } from "@/lib/api";

export async function listAvailableTaxesForUser(): Promise<ApiResult> {
  return apiFetch("/api/users/taxes/available");
}
