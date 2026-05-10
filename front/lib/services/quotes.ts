import { apiFetch, type ApiResult } from "@/lib/api";
import type {
  BackendQuote,
  BackendQuoteLine,
  BackendQuoteLineType,
} from "@/types/backend";

export type LineDraft = {
  type: BackendQuoteLineType;
  name: string;
  quantity: number;
  unit?: string;
  unitPriceEuros: number;
  position: number;
};

function toCents(euros: number): number {
  return Math.round(euros * 100);
}

function toLinePayload(draft: LineDraft) {
  return {
    type: draft.type,
    name: draft.name,
    quantity: String(draft.quantity),
    unit: draft.unit ?? "",
    unit_price: toCents(draft.unitPriceEuros),
    data: {},
    position: draft.position,
  };
}

export async function listQuotes(): Promise<ApiResult> {
  return apiFetch("/api/quotes");
}

export type CreateQuotePayload = {
  name: string;
  clientId: string;
  addressId: number;
};

export async function createQuote(
  payload: CreateQuotePayload,
): Promise<ApiResult> {
  return apiFetch("/api/quotes", {
    method: "POST",
    body: JSON.stringify({
      name: payload.name,
      client_id: payload.clientId,
      address_id: payload.addressId,
    }),
  });
}

export async function getQuote(quoteId: string): Promise<ApiResult> {
  return apiFetch(`/api/quotes/${encodeURIComponent(quoteId)}`);
}

export type UpdateQuotePayload = {
  name: string;
  clientId?: string;
  addressId?: number;
};

export async function updateQuote(
  quoteId: string,
  payload: UpdateQuotePayload,
): Promise<ApiResult> {
  const body: Record<string, unknown> = { name: payload.name };
  if (payload.clientId !== undefined) body.client_id = payload.clientId;
  if (payload.addressId !== undefined) body.address_id = payload.addressId;
  return apiFetch(`/api/quotes/${encodeURIComponent(quoteId)}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export async function dropQuote(quoteId: string): Promise<ApiResult> {
  return apiFetch(`/api/quotes/${encodeURIComponent(quoteId)}/drop`, {
    method: "POST",
  });
}

export async function continueQuote(quoteId: string): Promise<ApiResult> {
  return apiFetch(`/api/quotes/${encodeURIComponent(quoteId)}/continue`, {
    method: "POST",
  });
}

export async function createLine(
  quoteId: string,
  draft: LineDraft,
): Promise<ApiResult> {
  return apiFetch(`/api/quotes/${encodeURIComponent(quoteId)}/lines`, {
    method: "POST",
    body: JSON.stringify(toLinePayload(draft)),
  });
}

export async function updateLine(
  quoteId: string,
  lineId: string,
  draft: LineDraft,
): Promise<ApiResult> {
  return apiFetch(
    `/api/quotes/${encodeURIComponent(quoteId)}/lines/${encodeURIComponent(lineId)}`,
    {
      method: "PUT",
      body: JSON.stringify(toLinePayload(draft)),
    },
  );
}

export async function deleteLine(
  quoteId: string,
  lineId: string,
): Promise<ApiResult> {
  return apiFetch(
    `/api/quotes/${encodeURIComponent(quoteId)}/lines/${encodeURIComponent(lineId)}`,
    { method: "DELETE" },
  );
}

export function lineFromBackend(line: BackendQuoteLine): LineDraft & {
  lineId: string;
} {
  return {
    lineId: line.line_id,
    type: line.type,
    name: line.name,
    quantity: Number(line.quantity),
    unit: line.unit,
    unitPriceEuros: line.unit_price / 100,
    position: line.position,
  };
}

export type { BackendQuote, BackendQuoteLine };
