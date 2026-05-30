"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  getSchedule,
  updateScheduleCell,
  validateSchedule,
} from "@/lib/services/schedules";
import {
  type BackendScheduleDetails,
  scheduleBalanceState,
} from "@/types/backend";
import { formatEurosFromCents } from "@/lib/utils";
import { Button } from "@/components/ui/button";

type Params = {
  id: string;
};

export default function ScheduleDetailsPage() {
  const params = useParams<Params>();
  const scheduleId = params?.id ?? "";
  const [schedule, setSchedule] = useState<BackendScheduleDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const breadcrumbs = useMemo(
    () => [
      { href: "/schedule", label: "Échéanciers" },
      { href: `/schedule/${scheduleId}`, label: scheduleId || "Détail" },
    ],
    [scheduleId],
  );

  useEffect(() => {
    let cancelled = false;
    if (!scheduleId) return;
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
      setSchedule(body.schedule as BackendScheduleDetails);
      setError(null);
      setLoading(false);
    });
    return () => {
      cancelled = true;
    };
  }, [scheduleId]);

  async function onValidate() {
    if (!scheduleId) return;
    const { ok, body } = await validateSchedule(scheduleId);
    if (!ok || !body.success) {
      setError((body.message as string) ?? "Validation impossible.");
      return;
    }
    const refreshed = await getSchedule(scheduleId);
    if (refreshed.ok && refreshed.body.success && refreshed.body.schedule) {
      setSchedule(refreshed.body.schedule as BackendScheduleDetails);
      setError(null);
    }
  }

  async function setFirstLineToZero() {
    if (!schedule || schedule.lines.length === 0) return;
    const first = schedule.lines[0];
    const { ok, body } = await updateScheduleCell(schedule.schedule_id, {
      quoteLineId: first.quote_line_id,
      monthIndex: 1,
      amountEur: "0.00",
    });
    if (!ok || !body.success) {
      setError((body.message as string) ?? "Mise à jour impossible.");
      return;
    }
    const refreshed = await getSchedule(scheduleId);
    if (refreshed.ok && refreshed.body.success && refreshed.body.schedule) {
      setSchedule(refreshed.body.schedule as BackendScheduleDetails);
      setError(null);
    }
  }

  return (
    <>
      <PageBreadcrumb items={breadcrumbs} />
      <Card>
        <CardHeader className="flex flex-row items-center justify-between gap-4">
          <CardTitle>Échéancier {scheduleId}</CardTitle>
          <div className="flex gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={setFirstLineToZero}
            >
              Cellule #1 → 0,00
            </Button>
            <Button type="button" onClick={onValidate}>
              Valider l&apos;échéancier
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
                  <strong>Statut:</strong> {schedule.status}
                </p>
                <p>
                  <strong>Début:</strong> {schedule.start_month}
                </p>
                <p>
                  <strong>Durée:</strong> {schedule.duration_months} mois
                </p>
              </div>

              <div className="rounded-md border">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b bg-muted/40 text-left">
                      <th className="p-2">Ligne devis</th>
                      <th className="p-2">Planifié</th>
                      <th className="p-2">Attendu</th>
                      <th className="p-2">État</th>
                    </tr>
                  </thead>
                  <tbody>
                    {schedule.lines.map((line) => {
                      const state = scheduleBalanceState(
                        line.planned_cents,
                        line.expected_cents,
                      );
                      return (
                        <tr
                          key={line.quote_line_id}
                          className="border-b last:border-0"
                        >
                          <td className="p-2">{line.quote_line_id}</td>
                          <td className="p-2">
                            {formatEurosFromCents(line.planned_cents)}
                          </td>
                          <td className="p-2">
                            {formatEurosFromCents(line.expected_cents)}
                          </td>
                          <td className="p-2">{state}</td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
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
