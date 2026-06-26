"use client";

import PieChartCard from "@/components/charts/pie-chart-card";
import BarChartCard from "@/components/charts/bar-chart-card";
import type { BackendProjectQuoteRow } from "@/types/backend";

export const QUOTE_STATE_COLORS: Record<string, string> = {
  draft: "#94a3b8",
  negociation: "#f59e0b",
  validated: "#22c55e",
  drop: "#ef4444",
  sent: "#3b82f6",
};

export const QUOTE_STATE_LABELS: Record<string, string> = {
  draft: "Brouillon",
  negociation: "Négociation",
  validated: "Validé",
  drop: "Abandonné",
  sent: "Envoyé",
};

export const SCHEDULE_STATUS_COLORS: Record<string, string> = {
  DRAFT: "#94a3b8",
  NEGOCIATE: "#f59e0b",
  VALID: "#22c55e",
  DENIED: "#ef4444",
};

export const SCHEDULE_STATUS_LABELS: Record<string, string> = {
  DRAFT: "Brouillon",
  NEGOCIATE: "Négociation",
  VALID: "Validé",
  DENIED: "Refusé",
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
  const stateCount: Record<string, number> = {};
  for (const q of quotes) {
    stateCount[q.state] = (stateCount[q.state] ?? 0) + 1;
  }
  const statePieData = Object.entries(stateCount).map(([state, count]) => ({
    name: QUOTE_STATE_LABELS[state] ?? state,
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
    name: SCHEDULE_STATUS_LABELS[status] ?? status,
    count,
    color: SCHEDULE_STATUS_COLORS[status] ?? "#6b7280",
  }));

  const revenueData = [
    { name: "Total HT contractualisé", montant: totalHtCents / 100 },
    { name: "Encaissé", montant: collectedHtCents / 100 },
  ];

  if (quotes.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">Aucun devis rattaché au projet.</p>
    );
  }

  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
      <PieChartCard
        title="Répartition des devis par état"
        data={statePieData}
        outerRadius={70}
        height={180}
      />
      <BarChartCard
        title="Répartition des échéanciers"
        data={scheduleBarData}
        dataKey="count"
        colorKey="color"
        barName="Nb"
        height={180}
        noDataMessage="Aucun échéancier."
      />
      <BarChartCard
        title="Chiffre d'affaires (HT)"
        data={revenueData}
        dataKey="montant"
        defaultColor="#3b82f6"
        barName="Montant HT"
        height={180}
        tickFormatter={(v) => `${(v / 1000).toFixed(0)}k`}
        tooltipFormatter={(v) => formatEuros(v * 100)}
      />
    </div>
  );
}
