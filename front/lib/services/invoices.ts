import { apiFetch, type ApiResult } from "@/lib/api";
import type {
  BackendInvoiceDetails,
  BackendInvoiceSummary,
} from "@/types/backend";

export type CreateInvoiceFromSchedulePayload = {
  scheduleId: string;
  monthIndexes: number[];
  saleDate?: string;
  dueInDays?: number;
  issueNow?: boolean;
};

export type CreateInvoiceFromQuotePayload = {
  quoteId: string;
  saleDate?: string;
  dueInDays?: number;
  issueNow?: boolean;
};

export async function listInvoices(quoteId?: string): Promise<ApiResult> {
  const q = quoteId?.trim();
  const path = q
    ? `/api/invoices?quote_id=${encodeURIComponent(q)}`
    : "/api/invoices";
  return apiFetch(path);
}

export async function getInvoice(invoiceId: string): Promise<ApiResult> {
  return apiFetch(`/api/invoices/${encodeURIComponent(invoiceId)}`);
}

export async function createInvoiceFromSchedule(
  payload: CreateInvoiceFromSchedulePayload,
): Promise<ApiResult> {
  return apiFetch("/api/invoices/from-schedule", {
    method: "POST",
    body: JSON.stringify({
      schedule_id: payload.scheduleId,
      month_indexes: payload.monthIndexes,
      sale_date: payload.saleDate ?? "",
      due_in_days: payload.dueInDays ?? 0,
      issue_now: payload.issueNow ?? false,
    }),
  });
}

export async function createInvoiceFromQuote(
  payload: CreateInvoiceFromQuotePayload,
): Promise<ApiResult> {
  return apiFetch("/api/invoices/from-quote", {
    method: "POST",
    body: JSON.stringify({
      quote_id: payload.quoteId,
      sale_date: payload.saleDate ?? "",
      due_in_days: payload.dueInDays ?? 0,
      issue_now: payload.issueNow ?? false,
    }),
  });
}

export async function issueInvoice(invoiceId: string): Promise<ApiResult> {
  return apiFetch(`/api/invoices/${encodeURIComponent(invoiceId)}/issue`, {
    method: "POST",
  });
}

export async function cancelInvoice(invoiceId: string): Promise<ApiResult> {
  return apiFetch(`/api/invoices/${encodeURIComponent(invoiceId)}/cancel`, {
    method: "POST",
  });
}

export async function markInvoicePaid(invoiceId: string): Promise<ApiResult> {
  return apiFetch(`/api/invoices/${encodeURIComponent(invoiceId)}/paid`, {
    method: "POST",
  });
}

export function readInvoicesFromBody(
  body: Record<string, unknown>,
): BackendInvoiceSummary[] {
  if (!Array.isArray(body.invoices)) return [];
  return body.invoices as BackendInvoiceSummary[];
}

export function readInvoiceFromBody(
  body: Record<string, unknown>,
): BackendInvoiceDetails | null {
  if (!body.invoice || typeof body.invoice !== "object") return null;
  return body.invoice as BackendInvoiceDetails;
}
