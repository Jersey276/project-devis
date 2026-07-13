"use client";

import { useTranslations } from "next-intl";
import PieChartCard from "@/components/charts/pie-chart-card";
import BarChartCard from "@/components/charts/bar-chart-card";
import type { BackendProjectQuoteRow } from "@/types/backend";

export const QUOTE_STATE_COLORS: Record<string, string> = {
  draft: "#94a3b8",
  negociation: "#f59e0b",
  validated: "#22c55e",
  drop: "#ef4444",
  accepted: "#22c55e",
  refused: "#ef4444",
};

export const SCHEDULE_STATUS_COLORS: Record<string, string> = {
  DRAFT: "#94a3b8",
  NEGOCIATE: "#f59e0b",
  VALID: "#22c55e",
  DENIED: "#ef4444",
};

const euroFormatter = new Intl.NumberFormat("fr-FR", {
  style: "currency",
  currency: "EUR",
  maximumFractionDigits: 0,
});

export function formatEuros(cents: number): string {
  return euroFormatter.format(cents / 100);
}

type Props = {
  quotes: BackendProjectQuoteRow[];
  totalHtCents: number;
  collectedHtCents: number;
};

export default function ProjectCharts({ quotes, totalHtCents, collectedHtCents }: Props) {
  const t = useTranslations("project.detail.charts");
  const tTable = useTranslations("project.detail.quotesTable");
  const tQuoteStatus = useTranslations("status.quote");
  const tScheduleStatus = useTranslations("status.schedule");

  const stateCount: Record<string, number> = {};
  for (const q of quotes) {
    stateCount[q.state] = (stateCount[q.state] ?? 0) + 1;
  }
  const statePieData = Object.entries(stateCount).map(([state, count]) => ({
    name: tQuoteStatus(state),
    value: count,
    color: QUOTE_STATE_COLORS[state] ?? "#6b7280",
  }));

  const scheduleCount: Record<string, number> = {};
  for (const q of quotes) {
    for (const s of q.schedules ?? []) {
      scheduleCount[s.status] = (scheduleCount[s.status] ?? 0) + 1;
    }
  }
  const scheduleBarData = Object.entries(scheduleCount).map(([status, count]) => ({
    name: tScheduleStatus(status),
    count,
    color: SCHEDULE_STATUS_COLORS[status] ?? "#6b7280",
  }));

  const revenueData = [
    { name: t("totalHt"), montant: totalHtCents / 100 },
    { name: t("collected"), montant: collectedHtCents / 100 },
  ];

  if (quotes.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">{tTable("empty")}</p>
    );
  }

  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
      <PieChartCard
        title={t("quoteStates")}
        data={statePieData}
        outerRadius={70}
        height={180}
      />
      <BarChartCard
        title={t("scheduleStatuses")}
        data={scheduleBarData}
        dataKey="count"
        colorKey="color"
        barName={t("barNameCount")}
        height={180}
        noDataMessage={tTable("noSchedules")}
      />
      <BarChartCard
        title={t("revenue")}
        data={revenueData}
        dataKey="montant"
        defaultColor="#3b82f6"
        barName={t("barNameAmount")}
        height={180}
        tickFormatter={(v) => `${(v / 1000).toFixed(0)}k`}
        tooltipFormatter={(v) => formatEuros(v * 100)}
      />
    </div>
  );
}
