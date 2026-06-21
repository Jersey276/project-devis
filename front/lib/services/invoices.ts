import { apiFetch, type ApiResult } from "@/lib/api";
import type {
  BackendCreditNoteDetails,
  BackendCreditNoteSummary,
  BackendInvoiceDetails,
  BackendInvoiceLifecycleEvent,
  BackendInvoiceLifecycleStatus,
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

// OSS distance-selling threshold status for the current civil year (art. 259 D
// CGI): net B2C intra-EU turnover vs the EUR 10 000 threshold.
export async function getOSSThresholdStatus(): Promise<ApiResult> {
  return apiFetch("/api/invoices/oss-status");
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

export async function markInvoicePaid(invoiceId: string): Promise<ApiResult> {
  return apiFetch(`/api/invoices/${encodeURIComponent(invoiceId)}/paid`, {
    method: "POST",
  });
}

// Delete a DRAFT invoice. The backend refuses it once the invoice is issued
// (sealed/immutable), returning a finalized error.
export async function deleteDraftInvoice(invoiceId: string): Promise<ApiResult> {
  return apiFetch(`/api/invoices/${encodeURIComponent(invoiceId)}`, {
    method: "DELETE",
  });
}

// Advance the e-invoicing lifecycle status (réforme FR B2B). The backend is
// authoritative on the allowed transitions; only ISSUED/PAID invoices qualify.
export async function setInvoiceLifecycleStatus(
  invoiceId: string,
  status: Exclude<BackendInvoiceLifecycleStatus, "NONE">,
  note?: string,
): Promise<ApiResult> {
  return apiFetch(`/api/invoices/${encodeURIComponent(invoiceId)}/lifecycle`, {
    method: "POST",
    body: JSON.stringify({ status, note: note ?? "" }),
  });
}

// Deposit an issued invoice on the e-invoicing platform (B6). The backend
// drives the DEPOSITED lifecycle transition; a no-op platform is used until a
// PA (Plateforme Agréée) is contracted.
export async function depositInvoice(invoiceId: string): Promise<ApiResult> {
  return apiFetch(`/api/invoices/${encodeURIComponent(invoiceId)}/deposit`, {
    method: "POST",
  });
}

export async function listInvoiceLifecycleEvents(
  invoiceId: string,
): Promise<ApiResult> {
  return apiFetch(
    `/api/invoices/${encodeURIComponent(invoiceId)}/lifecycle-events`,
  );
}

export type CreateCreditNotePayload = {
  positions?: number[]; // empty/undefined = total credit of the remainder
  reason?: string;
};

export async function createCreditNote(
  invoiceId: string,
  payload: CreateCreditNotePayload = {},
): Promise<ApiResult> {
  return apiFetch(
    `/api/invoices/${encodeURIComponent(invoiceId)}/credit-notes`,
    {
      method: "POST",
      body: JSON.stringify({
        positions: payload.positions ?? [],
        reason: payload.reason ?? "",
      }),
    },
  );
}

export async function getCreditNote(creditNoteId: string): Promise<ApiResult> {
  return apiFetch(`/api/credit-notes/${encodeURIComponent(creditNoteId)}`);
}

export async function listCreditNotes(invoiceId?: string): Promise<ApiResult> {
  const i = invoiceId?.trim();
  const path = i
    ? `/api/credit-notes?invoice_id=${encodeURIComponent(i)}`
    : "/api/credit-notes";
  return apiFetch(path);
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

export function readLifecycleEventsFromBody(
  body: Record<string, unknown>,
): BackendInvoiceLifecycleEvent[] {
  if (!Array.isArray(body.events)) return [];
  return body.events as BackendInvoiceLifecycleEvent[];
}

export function readCreditNotesFromBody(
  body: Record<string, unknown>,
): BackendCreditNoteSummary[] {
  if (!Array.isArray(body.credit_notes)) return [];
  return body.credit_notes as BackendCreditNoteSummary[];
}

export function readCreditNoteFromBody(
  body: Record<string, unknown>,
): BackendCreditNoteDetails | null {
  if (!body.credit_note || typeof body.credit_note !== "object") return null;
  return body.credit_note as BackendCreditNoteDetails;
}
