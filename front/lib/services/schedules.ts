import { apiFetch, type ApiResult } from "@/lib/api";
import type {
  BackendScheduleDetails,
  BackendScheduleSummary,
} from "@/types/backend";

export type CreateSchedulePayload = {
  quoteId: string;
  name: string;
  startMonth: string;
  durationMonths: number;
};

export type UpdateScheduleCellPayload = {
  quoteLineId: string;
  monthIndex: number;
  amountEur: string;
};

export async function listSchedules(quoteId?: string): Promise<ApiResult> {
  const q = quoteId?.trim();
  const path = q
    ? `/api/schedules?quote_id=${encodeURIComponent(q)}`
    : "/api/schedules";
  return apiFetch(path);
}

export async function createSchedule(
  payload: CreateSchedulePayload,
): Promise<ApiResult> {
  return apiFetch("/api/schedules", {
    method: "POST",
    body: JSON.stringify({
      quote_id: payload.quoteId,
      name: payload.name,
      start_month: payload.startMonth,
      duration_months: payload.durationMonths,
    }),
  });
}

export async function getSchedule(scheduleId: string): Promise<ApiResult> {
  return apiFetch(`/api/schedules/${encodeURIComponent(scheduleId)}`);
}

export async function updateScheduleCell(
  scheduleId: string,
  payload: UpdateScheduleCellPayload,
): Promise<ApiResult> {
  return apiFetch(`/api/schedules/${encodeURIComponent(scheduleId)}/cells`, {
    method: "PATCH",
    body: JSON.stringify({
      quote_line_id: payload.quoteLineId,
      month_index: payload.monthIndex,
      amount_eur: payload.amountEur,
    }),
  });
}

export async function validateSchedule(scheduleId: string): Promise<ApiResult> {
  return apiFetch(`/api/schedules/${encodeURIComponent(scheduleId)}/validate`, {
    method: "POST",
  });
}

export function readSchedulesFromBody(
  body: Record<string, unknown>,
): BackendScheduleSummary[] {
  if (!Array.isArray(body.schedules)) return [];
  return body.schedules as BackendScheduleSummary[];
}

export function readScheduleFromBody(
  body: Record<string, unknown>,
): BackendScheduleDetails | null {
  if (!body.schedule || typeof body.schedule !== "object") return null;
  return body.schedule as BackendScheduleDetails;
}
