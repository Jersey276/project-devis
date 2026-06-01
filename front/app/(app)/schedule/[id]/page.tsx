"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { getSchedule, updateScheduleCell } from "@/lib/services/schedules";
import { exportSchedulePdf } from "@/lib/services/export";
import {
  type BackendScheduleDetails,
  type ScheduleBalanceState,
  scheduleBalanceState,
} from "@/types/backend";
import { cn, formatEurosFromCents } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { AlertCircle } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableFooter,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import ScheduleStatusSelect from "@/components/schedule/schedule-status-select";

type Params = {
  id: string;
};

type ScheduleMonthHeader = {
  year: string;
  label: string;
};

type ScheduleYearHeader = {
  year: string;
  span: number;
};

function centsToEuros(cents: number): string {
  return (cents / 100).toFixed(2);
}

function parseScheduleMonth(month: string): Date {
  const [yearPart, monthPart] = month.split("-");
  const year = Number.parseInt(yearPart ?? "", 10);
  const monthIndex = Number.parseInt(monthPart ?? "", 10);
  return new Date(Date.UTC(year, monthIndex - 1, 1));
}

function buildScheduleMonthHeaders(
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

function buildScheduleYearHeaders(
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

function draftKey(lineId: string, monthIndex: number): string {
  return `${lineId}::${monthIndex}`;
}

function buildCellDrafts(
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
    drafts[draftKey(line.quote_line_id, 1)] = centsToEuros(line.planned_cents);
  }
  return drafts;
}

function balanceStateClasses(state: ScheduleBalanceState): {
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

function eurosStringToCents(value: string): number {
  return Math.round(Number.parseFloat(value) * 100);
}

function applyOptimisticCellUpdate(
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

type ParsedAmountResult =
  | { ok: true; normalizedValue: string }
  | { ok: false; reason: "empty" | "invalid"; message?: string };

function parseAmountEur(value: string): ParsedAmountResult {
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

export default function ScheduleDetailsPage() {
  const params = useParams<Params>();
  const scheduleId = params?.id ?? "";
  const [schedule, setSchedule] = useState<BackendScheduleDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [cellDrafts, setCellDrafts] = useState<Record<string, string>>({});
  const [savedCellDrafts, setSavedCellDrafts] = useState<
    Record<string, string>
  >({});
  const [cellErrors, setCellErrors] = useState<Record<string, string>>({});
  const [isExporting, setIsExporting] = useState(false);

  const breadcrumbs = useMemo(
    () => [
      { href: "/schedule", label: "Échéanciers" },
      { href: `/schedule/${scheduleId}`, label: scheduleId || "Détail" },
    ],
    [scheduleId],
  );

  const monthHeaders = useMemo(
    () =>
      schedule
        ? buildScheduleMonthHeaders(
            schedule.start_month,
            schedule.duration_months,
          )
        : [],
    [schedule],
  );

  const yearHeaders = useMemo(
    () => buildScheduleYearHeaders(monthHeaders),
    [monthHeaders],
  );

  const columnTotalsByMonth = useMemo(() => {
    if (!schedule) return new Map<number, number>();
    return new Map(
      schedule.column_totals.map((column) => [
        column.month_index,
        column.amount_cents,
      ]),
    );
  }, [schedule]);

  const isReadOnly =
    schedule?.status === "VALID" || schedule?.status === "DENIED";

  const loadSchedule = useCallback(async () => {
    if (!scheduleId) return;
    const { ok, body } = await getSchedule(scheduleId);
    if (!ok || !body.success || !body.schedule) {
      setError(
        (body.message as string) ?? "Impossible de charger l'échéancier.",
      );
      setSchedule(null);
      setLoading(false);
      return;
    }
    const loadedSchedule = body.schedule as BackendScheduleDetails;
    const drafts = buildCellDrafts(loadedSchedule);
    setSchedule(loadedSchedule);
    setCellDrafts(drafts);
    setSavedCellDrafts(drafts);
    setCellErrors({});
    setError(null);
    setLoading(false);
  }, [scheduleId]);

  useEffect(() => {
    let cancelled = false;
    if (!scheduleId) return;
    loadSchedule().then(() => {
      if (cancelled) return;
    });
    return () => {
      cancelled = true;
    };
  }, [loadSchedule, scheduleId]);

  async function onExportPdf() {
    if (!scheduleId || isExporting) return;

    setIsExporting(true);
    try {
      await exportSchedulePdf(scheduleId);
      setError(null);
    } catch {
      setError("Export PDF impossible.");
    } finally {
      setIsExporting(false);
    }
  }

  function onCellDraftChange(
    quoteLineId: string,
    monthIndex: number,
    value: string,
  ) {
    const key = draftKey(quoteLineId, monthIndex);
    setCellDrafts((prev) => ({
      ...prev,
      [key]: value,
    }));
    setCellErrors((prev) => {
      if (!prev[key]) return prev;
      const next = { ...prev };
      delete next[key];
      return next;
    });
  }

  function focusCellInput(lineIndex: number, monthIndex: number) {
    if (!schedule) return;
    const line = schedule.lines[lineIndex];
    if (!line) return;
    const input = document.querySelector<HTMLInputElement>(
      `input[name='cell-${line.quote_line_id}-m${monthIndex}']`,
    );
    input?.focus();
    input?.select();
  }

  function handleCellKeyDown(
    e: React.KeyboardEvent<HTMLInputElement>,
    lineIndex: number,
    monthIndex: number,
  ) {
    if (!schedule || isReadOnly) return;

    const maxLineIndex = schedule.lines.length - 1;
    const maxMonthIndex = schedule.duration_months;

    if (e.key === "Enter") {
      e.preventDefault();
      e.currentTarget.blur();
      if (lineIndex < maxLineIndex) {
        setTimeout(() => {
          focusCellInput(lineIndex + 1, monthIndex);
        }, 0);
      }
      return;
    }

    if (e.key === "Tab") {
      e.preventDefault();

      if (e.shiftKey) {
        if (monthIndex > 1) {
          focusCellInput(lineIndex, monthIndex - 1);
          return;
        }
        if (lineIndex > 0) {
          focusCellInput(lineIndex - 1, maxMonthIndex);
        }
        return;
      }

      if (monthIndex < maxMonthIndex) {
        focusCellInput(lineIndex, monthIndex + 1);
        return;
      }
      if (lineIndex < maxLineIndex) {
        focusCellInput(lineIndex + 1, 1);
      }
      return;
    }

    if (e.key === "ArrowRight") {
      e.preventDefault();
      if (monthIndex < maxMonthIndex) {
        focusCellInput(lineIndex, monthIndex + 1);
      }
      return;
    }

    if (e.key === "ArrowLeft") {
      e.preventDefault();
      if (monthIndex > 1) {
        focusCellInput(lineIndex, monthIndex - 1);
      }
      return;
    }

    if (e.key === "ArrowDown") {
      e.preventDefault();
      if (lineIndex < maxLineIndex) {
        focusCellInput(lineIndex + 1, monthIndex);
      }
      return;
    }

    if (e.key === "ArrowUp") {
      e.preventDefault();
      if (lineIndex > 0) {
        focusCellInput(lineIndex - 1, monthIndex);
      }
    }
  }

  async function saveCell(quoteLineId: string, monthIndex: number) {
    if (!schedule || isReadOnly) return;

    const key = draftKey(quoteLineId, monthIndex);
    const rawAmount = cellDrafts[key] ?? "";

    const parsedAmount = parseAmountEur(rawAmount);
    if (!parsedAmount.ok) {
      if (parsedAmount.reason === "empty") {
        setCellDrafts((prev) => ({
          ...prev,
          [key]: savedCellDrafts[key] ?? "",
        }));
        setCellErrors((prev) => {
          if (!prev[key]) return prev;
          const next = { ...prev };
          delete next[key];
          return next;
        });
        setError(null);
        return;
      }

      setCellErrors((prev) => ({
        ...prev,
        [key]: parsedAmount.message ?? "Montant invalide.",
      }));
      setError(null);
      return;
    }

    const savedParsed = parseAmountEur(savedCellDrafts[key] ?? "");
    if (
      savedParsed.ok &&
      savedParsed.normalizedValue === parsedAmount.normalizedValue
    ) {
      setCellDrafts((prev) => ({
        ...prev,
        [key]: parsedAmount.normalizedValue,
      }));
      setCellErrors((prev) => {
        if (!prev[key]) return prev;
        const next = { ...prev };
        delete next[key];
        return next;
      });
      setError(null);
      return;
    }

    const { ok, body } = await updateScheduleCell(schedule.schedule_id, {
      quoteLineId,
      monthIndex,
      amountEur: parsedAmount.normalizedValue,
    });
    if (!ok || !body.success) {
      const message = (body.message as string) ?? "Mise à jour impossible.";
      setCellErrors((prev) => ({
        ...prev,
        [key]: message,
      }));
      setError(null);
      return;
    }

    const previousAmountCents = savedParsed.ok
      ? eurosStringToCents(savedParsed.normalizedValue)
      : 0;
    const amountCents = eurosStringToCents(parsedAmount.normalizedValue);

    setSchedule((prev) =>
      prev
        ? applyOptimisticCellUpdate(
            prev,
            quoteLineId,
            monthIndex,
            amountCents,
            previousAmountCents,
          )
        : prev,
    );
    setCellDrafts((prev) => ({
      ...prev,
      [key]: parsedAmount.normalizedValue,
    }));
    setSavedCellDrafts((prev) => ({
      ...prev,
      [key]: parsedAmount.normalizedValue,
    }));
    setCellErrors((prev) => {
      if (!prev[key]) return prev;
      const next = { ...prev };
      delete next[key];
      return next;
    });
    setError(null);
  }

  return (
    <>
      <PageBreadcrumb items={breadcrumbs} />
      <Card>
        <CardHeader className="flex flex-row items-center justify-between gap-4">
          <CardTitle>Échéancier {scheduleId}</CardTitle>
          <div className="flex items-center gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={onExportPdf}
              disabled={isExporting || loading || !scheduleId}
            >
              {isExporting ? "Export..." : "Exporter PDF"}
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {!scheduleId ? (
            <p className="text-destructive">
              Identifiant d&apos;échéancier invalide.
            </p>
          ) : null}
          {loading ? <p>Chargement…</p> : null}
          {!loading && error ? (
            <p className="text-destructive">{error}</p>
          ) : null}
          {!loading && !error && schedule ? (
            <div className="space-y-4">
              <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-4">
                <p>
                  <strong>Nom:</strong> {schedule.name}
                </p>
                <p>
                  <strong>Statut:</strong>{" "}
                  <ScheduleStatusSelect
                    scheduleId={schedule.schedule_id}
                    value={schedule.status}
                    className="w-44"
                    onUpdated={loadSchedule}
                    onError={setError}
                  />
                </p>
                <p>
                  <strong>Début:</strong> {schedule.start_month}
                </p>
                <p>
                  <strong>Durée:</strong> {schedule.duration_months} mois
                </p>
              </div>

              <div className="w-full max-w-full min-w-0 overflow-x-hidden rounded-md border">
                <Table
                  className="min-w-max"
                  containerClassName="max-w-full overflow-x-auto overflow-y-hidden"
                >
                  <TableHeader className="bg-muted/40">
                    <TableRow className="text-left">
                      <TableHead
                        rowSpan={2}
                        className="sticky left-0 z-30 min-w-44 bg-muted/40"
                      >
                        Ligne devis
                      </TableHead>
                      <TableHead
                        rowSpan={2}
                        className="sticky left-44 z-30 min-w-32 bg-muted/40 text-center"
                      >
                        Total
                      </TableHead>
                      <TableHead
                        rowSpan={2}
                        className="sticky left-76 z-30 min-w-32 bg-muted/40 text-center"
                      >
                        Restant
                      </TableHead>
                      {yearHeaders.map((header) => (
                        <TableHead
                          key={`head-year-${header.year}`}
                          colSpan={header.span}
                          className="text-center"
                        >
                          {header.year}
                        </TableHead>
                      ))}
                    </TableRow>
                    <TableRow className="text-left">
                      {monthHeaders.map((header, index) => (
                        <TableHead
                          key={`head-month-${index + 1}`}
                          className="text-center uppercase"
                        >
                          {header.label}
                        </TableHead>
                      ))}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {schedule.lines.map((line, lineIndex) => {
                      const state = scheduleBalanceState(
                        line.planned_cents,
                        line.expected_cents,
                      );
                      const stateClasses = balanceStateClasses(state);
                      return (
                        <TableRow
                          key={line.quote_line_id}
                          className={stateClasses.rowClass}
                        >
                          <TableCell
                            className={cn(
                              "sticky left-0 z-20 min-w-44",
                              stateClasses.stickyCellClass,
                            )}
                          >
                            {line.quote_line_id}
                          </TableCell>
                          <TableCell
                            data-testid={`line-total-${line.quote_line_id}`}
                            className={cn(
                              "sticky left-44 z-20 min-w-32 text-center",
                              stateClasses.stickyCellClass,
                            )}
                          >
                            {formatEurosFromCents(line.expected_cents)}
                          </TableCell>
                          <TableCell
                            data-testid={`line-remaining-${line.quote_line_id}`}
                            className={cn(
                              "sticky left-76 z-20 min-w-32 text-center",
                              stateClasses.stickyCellClass,
                            )}
                          >
                            {formatEurosFromCents(
                              line.expected_cents - line.planned_cents,
                            )}
                          </TableCell>
                          {monthHeaders.map((_, index) => {
                            const monthIndex = index + 1;
                            const key = draftKey(
                              line.quote_line_id,
                              monthIndex,
                            );
                            const cellError = cellErrors[key];
                            return (
                              <TableCell
                                key={`${line.quote_line_id}-month-${monthIndex}`}
                              >
                                <div className="relative inline-flex items-center">
                                  <input
                                    name={`cell-${line.quote_line_id}-m${monthIndex}`}
                                    className={cn(
                                      "h-9 w-24 rounded-md border px-2",
                                      cellError
                                        ? "border-destructive pr-8 focus-visible:ring-destructive"
                                        : null,
                                    )}
                                    aria-invalid={Boolean(cellError)}
                                    disabled={isReadOnly}
                                    value={cellDrafts[key] ?? ""}
                                    onChange={(e) =>
                                      onCellDraftChange(
                                        line.quote_line_id,
                                        monthIndex,
                                        e.target.value,
                                      )
                                    }
                                    onBlur={() => {
                                      void saveCell(
                                        line.quote_line_id,
                                        monthIndex,
                                      );
                                    }}
                                    onKeyDown={(e) => {
                                      handleCellKeyDown(
                                        e,
                                        lineIndex,
                                        monthIndex,
                                      );
                                    }}
                                  />
                                  {cellError ? (
                                    <Tooltip>
                                      <TooltipTrigger asChild>
                                        <button
                                          type="button"
                                          data-testid={`cell-error-${line.quote_line_id}-m${monthIndex}`}
                                          className="text-destructive absolute right-2 inline-flex h-4 w-4 items-center justify-center"
                                          aria-label={cellError}
                                          title={cellError}
                                        >
                                          <AlertCircle className="h-4 w-4" />
                                        </button>
                                      </TooltipTrigger>
                                      <TooltipContent side="top" sideOffset={6}>
                                        {cellError}
                                      </TooltipContent>
                                    </Tooltip>
                                  ) : null}
                                </div>
                              </TableCell>
                            );
                          })}
                        </TableRow>
                      );
                    })}
                  </TableBody>
                  <TableFooter>
                    <TableRow>
                      <TableCell className="sticky left-0 z-20 min-w-44 bg-muted/60 font-semibold">
                        Totaux mensuels
                      </TableCell>
                      <TableCell
                        data-testid="footer-total-quote"
                        className="sticky left-44 z-20 min-w-32 bg-muted/60 text-center font-semibold"
                      >
                        {formatEurosFromCents(schedule.quote_total_cents)}
                      </TableCell>
                      <TableCell
                        data-testid="footer-total-remaining"
                        className="sticky left-76 z-20 min-w-32 bg-muted/60 text-center font-semibold"
                      >
                        {formatEurosFromCents(
                          schedule.quote_total_cents -
                            schedule.planned_total_cents,
                        )}
                      </TableCell>
                      {monthHeaders.map((_, index) => {
                        const monthIndex = index + 1;
                        return (
                          <TableCell
                            key={`footer-month-total-${monthIndex}`}
                            data-testid={`footer-month-total-${monthIndex}`}
                            className="text-center font-semibold"
                          >
                            {formatEurosFromCents(
                              columnTotalsByMonth.get(monthIndex) ?? 0,
                            )}
                          </TableCell>
                        );
                      })}
                    </TableRow>
                  </TableFooter>
                </Table>
              </div>

              <p>
                <strong>Total planifié:</strong>{" "}
                {formatEurosFromCents(schedule.planned_total_cents)}
              </p>
              <p>
                <strong>Total devis:</strong>{" "}
                {formatEurosFromCents(schedule.quote_total_cents)}
              </p>
            </div>
          ) : null}
        </CardContent>
      </Card>
    </>
  );
}
