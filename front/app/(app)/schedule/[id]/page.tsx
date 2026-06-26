"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { getSchedule } from "@/lib/services/schedules";
import { exportSchedulePdf } from "@/lib/services/export";
import GenerateInvoiceFromScheduleDialog from "@/components/invoice/generate-invoice-from-schedule-dialog";
import { type BackendScheduleDetails } from "@/types/backend";
import { formatEurosFromCents } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { useMode } from "@/lib/mode-context";
import {
  Table,
  TableBody,
} from "@/components/ui/table";
import { buildCellDrafts } from "./schedule-utils";
import { useScheduleTableStructure } from "./use-schedule-table-structure";
import { useScheduleCellEditing } from "./use-schedule-cell-editing";
import ScheduleMetadata from "./schedule-metadata";
import ScheduleTableHeader from "./schedule-table-header";
import ScheduleFooter from "./schedule-footer";
import ScheduleLineRow from "./schedule-line-row";

type Params = {
  id: string;
};

export default function ScheduleDetailsPage() {
  const { isCustomer } = useMode();
  const params = useParams<Params>();
  const scheduleId = params?.id ?? "";
  const [schedule, setSchedule] = useState<BackendScheduleDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isExporting, setIsExporting] = useState(false);
  const [invoiceDialogOpen, setInvoiceDialogOpen] = useState(false);

  const breadcrumbs = useMemo(
    () => [
      { href: "/schedule", label: "Échéanciers" },
      {
        href: `/schedule/${scheduleId}`,
        label: schedule?.name || scheduleId || "Détail",
      },
    ],
    [scheduleId, schedule?.name],
  );

  const {
    monthHeaders,
    yearHeaders,
    columnTotalsByMonth,
    sortedLines,
    editableLines,
    editableIndexByLineId,
  } = useScheduleTableStructure(schedule);

  const isReadOnly =
    isCustomer || schedule?.status === "VALID" || schedule?.status === "DENIED";

  const {
    cellDrafts,
    setCellDrafts,
    setSavedCellDrafts,
    cellErrors,
    setCellErrors,
    onCellDraftChange,
    handleCellKeyDown,
    saveCell,
  } = useScheduleCellEditing({
    schedule,
    setSchedule,
    editableLines,
    editableIndexByLineId,
    isReadOnly,
    setError,
  });

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
  }, [scheduleId, setCellDrafts, setSavedCellDrafts, setCellErrors]);

  useEffect(() => {
    if (!scheduleId) return;
    let cancelled = false;
    getSchedule(scheduleId).then(({ ok, body }) => {
      if (cancelled) return;
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
    });
    return () => {
      cancelled = true;
    };
  }, [scheduleId, setCellDrafts, setSavedCellDrafts, setCellErrors]);

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

  return (
    <>
      <PageBreadcrumb items={breadcrumbs} />
      <Card>
        <CardHeader className="flex flex-row items-center justify-between gap-4">
          <CardTitle>Échéancier {schedule?.name}</CardTitle>
          <div className="flex items-center gap-2">
            {!isCustomer && schedule?.status === "VALID" ? (
              <Button
                type="button"
                onClick={() => setInvoiceDialogOpen(true)}
                disabled={loading || !scheduleId}
              >
                Générer une facture
              </Button>
            ) : null}
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
              <ScheduleMetadata
                schedule={schedule}
                isCustomer={isCustomer}
                onUpdated={loadSchedule}
                onError={setError}
              />

              <div className="w-full max-w-full min-w-0 overflow-x-hidden rounded-md border">
                <Table
                  className="min-w-max"
                  containerClassName="max-w-full overflow-x-auto overflow-y-hidden"
                >
                  <ScheduleTableHeader
                    yearHeaders={yearHeaders}
                    monthHeaders={monthHeaders}
                  />
                  <TableBody>
                    {sortedLines.map((line) => (
                      <ScheduleLineRow
                        key={line.quote_line_id}
                        line={line}
                        monthHeaders={monthHeaders}
                        editableIndexByLineId={editableIndexByLineId}
                        cellDrafts={cellDrafts}
                        cellErrors={cellErrors}
                        isReadOnly={isReadOnly}
                        onCellDraftChange={onCellDraftChange}
                        onCellBlur={(id, idx) => void saveCell(id, idx)}
                        onCellKeyDown={handleCellKeyDown}
                      />
                    ))}
                  </TableBody>
                  <ScheduleFooter
                    schedule={schedule}
                    monthHeaders={monthHeaders}
                    columnTotalsByMonth={columnTotalsByMonth}
                  />
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

      {!isCustomer && schedule ? (
        <GenerateInvoiceFromScheduleDialog
          open={invoiceDialogOpen}
          onOpenChange={setInvoiceDialogOpen}
          scheduleId={scheduleId}
          durationMonths={schedule.duration_months}
          columnTotals={schedule.column_totals}
        />
      ) : null}
    </>
  );
}
