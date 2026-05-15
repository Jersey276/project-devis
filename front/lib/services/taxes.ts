import { apiFetch, type ApiResult } from "@/lib/api";

export async function listAvailableTaxesForUser(
  includeIds: number[] = [],
): Promise<ApiResult> {
  const qs =
    includeIds.length > 0 ? `?include_ids=${includeIds.join(",")}` : "";
  return apiFetch(`/api/users/taxes/available${qs}`);
}
