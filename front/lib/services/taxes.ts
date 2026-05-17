import { apiFetch, type ApiResult } from "@/lib/api";

export async function listAvailableTaxesForUser(
  includeIds: number[] = [],
  addressId?: number,
): Promise<ApiResult> {
  const params = new URLSearchParams();
  if (includeIds.length > 0) params.set("include_ids", includeIds.join(","));
  if (addressId !== undefined && addressId > 0)
    params.set("address_id", String(addressId));
  const qs = params.toString();
  return apiFetch(`/api/users/taxes/available${qs ? `?${qs}` : ""}`);
}
