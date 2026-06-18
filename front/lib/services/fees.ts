import { apiFetch, type ApiResult } from "@/lib/api";
import type { BackendFee, FeeCategory } from "@/types/backend";

// Full-replace shape sent to the gateway. unit_price is in cents; tax_id 0 (or
// omitted) means "no tax".
export type FeePayload = {
  category: FeeCategory;
  name: string;
  unit: string;
  unit_price: number;
  tax_id: number;
};

export async function listFees(
  includeArchived = false,
): Promise<ApiResult> {
  const qs = includeArchived ? "?archived=true" : "";
  return apiFetch(`/api/fees${qs}`);
}

export async function createFee(payload: FeePayload): Promise<ApiResult> {
  return apiFetch("/api/fees", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function getFee(feeId: string): Promise<ApiResult> {
  return apiFetch(`/api/fees/${encodeURIComponent(feeId)}`);
}

export async function updateFee(
  feeId: string,
  payload: FeePayload,
): Promise<ApiResult> {
  return apiFetch(`/api/fees/${encodeURIComponent(feeId)}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export async function archiveFee(feeId: string): Promise<ApiResult> {
  return apiFetch(`/api/fees/${encodeURIComponent(feeId)}`, {
    method: "DELETE",
  });
}

export type { BackendFee };
