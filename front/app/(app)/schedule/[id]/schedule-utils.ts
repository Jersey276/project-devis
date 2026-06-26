import type { BackendScheduleDetails, ScheduleBalanceState } from "@/types/backend";

export type ScheduleMonthHeader = {
  year: string;
  label: string;
};

export type ScheduleYearHeader = {
  year: string;
  span: number;
};

export type ParsedAmountResult =
  | { ok: true; normalizedValue: string }
  | { ok: false; reason: "empty" | "invalid"; message?: string };

export function centsToEuros(cents: number): string {
  return (cents / 100).toFixed(2);
}

export function parseScheduleMonth(month: string): Date {
  const [yearPart, monthPart] = month.split("-");
  const year = Number.parseInt(yearPart ?? "", 10);
  const monthIndex = Number.parseInt(monthPart ?? "", 10);
  return new Date(Date.UTC(year, monthIndex - 1, 1));
}

export function buildScheduleMonthHeaders(
  startMonth: string,
  durationMonths: number,
): ScheduleMonthHeader[] {
  const headers: ScheduleMonthHeader[] = [];
  const current = parseScheduleMonth(startMonth);

  for (let monthIndex = 0; monthIndex < durationMonths; monthIndex += 1) {
    headers.push({
      year: String(current.getUTCFullYear()),
      label: current.toLocaleString("fr-FR", {
        month: "long",
        timeZone: "UTC",
      }),
    });
    current.setUTCMonth(current.getUTCMonth() + 1);
  }

  return headers;
}

export function buildScheduleYearHeaders(
  monthHeaders: ScheduleMonthHeader[],
): ScheduleYearHeader[] {
  const headers: ScheduleYearHeader[] = [];

  for (const monthHeader of monthHeaders) {
    const lastHeader = headers.at(-1);
    if (lastHeader?.year === monthHeader.year) {
      lastHeader.span += 1;
      continue;
    }

    headers.push({ year: monthHeader.year, span: 1 });
  }

  return headers;
}

export function draftKey(lineId: string, monthIndex: number): string {
  return `${lineId}::${monthIndex}`;
}

export function buildCellDrafts(
  schedule: BackendScheduleDetails,
): Record<string, string> {
  const drafts: Record<string, string> = {};

  if (Array.isArray(schedule.cells) && schedule.cells.length > 0) {
    for (const cell of schedule.cells) {
      drafts[draftKey(cell.quote_line_id, cell.month_index)] = centsToEuros(
        cell.amount_cents,
      );
    }
    return drafts;
  }

  // Backward-compatible fallback while details payload does not provide cell-level data.
  for (const line of schedule.lines) {
    if (line.data_kind === "group") continue;
    drafts[draftKey(line.quote_line_id, 1)] = centsToEuros(line.planned_cents);
  }
  return drafts;
}

export function balanceStateClasses(state: ScheduleBalanceState): {
  rowClass: string;
  stickyCellClass: string;
} {
  switch (state) {
    case "under":
      return {
        rowClass: "bg-amber-50/60 hover:bg-amber-100/60",
        stickyCellClass: "bg-amber-50/60",
      };
    case "over":
      return {
        rowClass: "bg-rose-50/60 hover:bg-rose-100/60",
        stickyCellClass: "bg-rose-50/60",
      };
    default:
      return {
        rowClass: "bg-emerald-50/60 hover:bg-emerald-100/60",
        stickyCellClass: "bg-emerald-50/60",
      };
  }
}

export function eurosStringToCents(value: string): number {
  return Math.round(Number.parseFloat(value) * 100);
}

export function applyOptimisticCellUpdate(
  schedule: BackendScheduleDetails,
  quoteLineId: string,
  monthIndex: number,
  amountCents: number,
  previousAmountCents: number,
): BackendScheduleDetails {
  const delta = amountCents - previousAmountCents;
  if (delta === 0) return schedule;

  const lines = schedule.lines.map((line) =>
    line.quote_line_id === quoteLineId
      ? { ...line, planned_cents: line.planned_cents + delta }
      : line,
  );

  const existingCells = Array.isArray(schedule.cells) ? schedule.cells : [];
  const targetCellIndex = existingCells.findIndex(
    (cell) =>
      cell.quote_line_id === quoteLineId && cell.month_index === monthIndex,
  );
  const cells =
    targetCellIndex >= 0
      ? existingCells.map((cell, index) =>
          index === targetCellIndex
            ? { ...cell, amount_cents: amountCents }
            : cell,
        )
      : [
          ...existingCells,
          {
            quote_line_id: quoteLineId,
            month_index: monthIndex,
            amount_cents: amountCents,
          },
        ];

  const targetColumnIndex = schedule.column_totals.findIndex(
    (column) => column.month_index === monthIndex,
  );
  const column_totals =
    targetColumnIndex >= 0
      ? schedule.column_totals.map((column, index) =>
          index === targetColumnIndex
            ? { ...column, amount_cents: column.amount_cents + delta }
            : column,
        )
      : [
          ...schedule.column_totals,
          {
            month_index: monthIndex,
            amount_cents: delta,
          },
        ];

  return {
    ...schedule,
    lines,
    cells,
    column_totals,
    planned_total_cents: schedule.planned_total_cents + delta,
  };
}

export function parseAmountEur(value: string): ParsedAmountResult {
  const sanitized = value.trim().replace(",", ".");
  if (!sanitized) return { ok: false, reason: "empty" };
  if (!/^\d+(\.\d{1,2})?$/.test(sanitized)) {
    return {
      ok: false,
      reason: "invalid",
      message:
        "Montant invalide. Utilisez un nombre positif avec 2 décimales max.",
    };
  }

  const amount = Number.parseFloat(sanitized);
  if (!Number.isFinite(amount) || amount < 0) {
    return {
      ok: false,
      reason: "invalid",
      message:
        "Montant invalide. Utilisez un nombre positif avec 2 décimales max.",
    };
  }

  return { ok: true, normalizedValue: amount.toFixed(2) };
}
