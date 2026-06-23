"use client";

import {
  PieChart,
  Pie,
  Cell,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";
import type { BackendProjectQuoteRow } from "@/types/backend";

const QUOTE_STATE_COLORS: Record<string, string> = {
  draft: "#94a3b8",
  negociation: "#f59e0b",
  validated: "#22c55e",
  drop: "#ef4444",
  sent: "#3b82f6",
};

const QUOTE_STATE_LABELS: Record<string, string> = {
  draft: "Brouillon",
  negociation: "Négociation",
  validated: "Validé",
  drop: "Abandonné",
  sent: "Envoyé",
};

const SCHEDULE_STATUS_COLORS: Record<string, string> = {
  DRAFT: "#94a3b8",
  NEGOCIATE: "#f59e0b",
  VALID: "#22c55e",
  DENIED: "#ef4444",
};

const SCHEDULE_STATUS_LABELS: Record<string, string> = {
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

function formatEuros(cents: number): string {
  return euroFormatter.format(cents / 100);
}

type Props = {
  quotes: BackendProjectQuoteRow[];
  totalHtCents: number;
  collectedHtCents: number;
};

export default function ProjectCharts({ quotes, totalHtCents, collectedHtCents }: Props) {
  // Quote state distribution
  const stateCount: Record<string, number> = {};
  for (const q of quotes) {
    stateCount[q.state] = (stateCount[q.state] ?? 0) + 1;
  }
  const statePieData = Object.entries(stateCount).map(([state, count]) => ({
    name: QUOTE_STATE_LABELS[state] ?? state,
    value: count,
    color: QUOTE_STATE_COLORS[state] ?? "#6b7280",
  }));

  // Schedule status distribution
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

  // Revenue chart
  const revenueData = [
    { name: "Total HT contractualisé", montant: totalHtCents / 100 },
    { name: "Encaissé", montant: collectedHtCents / 100 },
  ];

  const hasQuotes = quotes.length > 0;
  const hasSchedules = scheduleBarData.length > 0;

  if (!hasQuotes) {
    return (
      <p className="text-sm text-muted-foreground">Aucun devis rattaché au projet.</p>
    );
  }

  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
      {/* Pie — quote states */}
      <div className="rounded-lg border bg-card p-4">
        <p className="mb-3 text-sm font-medium">Répartition des devis par état</p>
        <ResponsiveContainer width="100%" height={180}>
          <PieChart>
            <Pie
              data={statePieData}
              cx="50%"
              cy="50%"
              outerRadius={70}
              dataKey="value"
              label={({ name, value }) => `${name} (${value})`}
              labelLine={false}
            >
              {statePieData.map((entry, i) => (
                <Cell key={i} fill={entry.color} />
              ))}
            </Pie>
            <Tooltip formatter={(v) => [`${v}`, "Nb"]} />
          </PieChart>
        </ResponsiveContainer>
      </div>

      {/* Bar — schedule statuses */}
      <div className="rounded-lg border bg-card p-4">
        <p className="mb-3 text-sm font-medium">Répartition des échéanciers</p>
        {hasSchedules ? (
          <ResponsiveContainer width="100%" height={180}>
            <BarChart data={scheduleBarData} margin={{ top: 4, right: 4, left: -20, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" tick={{ fontSize: 11 }} />
              <YAxis allowDecimals={false} tick={{ fontSize: 11 }} />
              <Tooltip />
              <Bar dataKey="count" name="Nb">
                {scheduleBarData.map((entry, i) => (
                  <Cell key={i} fill={entry.color} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        ) : (
          <p className="text-sm text-muted-foreground mt-4">Aucun échéancier.</p>
        )}
      </div>

      {/* Bar — revenue */}
      <div className="rounded-lg border bg-card p-4">
        <p className="mb-3 text-sm font-medium">Chiffre d&apos;affaires (HT)</p>
        <ResponsiveContainer width="100%" height={180}>
          <BarChart data={revenueData} margin={{ top: 4, right: 4, left: 0, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="name" tick={{ fontSize: 10 }} />
            <YAxis tick={{ fontSize: 11 }} tickFormatter={(v) => `${(v / 1000).toFixed(0)}k`} />
            <Tooltip formatter={(v) => [formatEuros((v as number) * 100), "Montant HT"]} />
            <Bar dataKey="montant" fill="#3b82f6" name="Montant HT" />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
