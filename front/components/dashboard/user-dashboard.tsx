"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import Link from "next/link";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import PieChartCard from "@/components/charts/pie-chart-card";
import BarChartCard from "@/components/charts/bar-chart-card";
import {
  QUOTE_STATE_COLORS,
  QUOTE_STATE_LABELS,
  SCHEDULE_STATUS_COLORS,
  SCHEDULE_STATUS_LABELS,
  formatEuros,
} from "@/components/project/project-charts";
import { listProjects, getProjectDetail } from "@/lib/services/projects";
import { listQuotes } from "@/lib/services/quotes";
import type {
  BackendProject,
  BackendQuote,
  BackendProjectDetail,
} from "@/types/backend";

const QUOTE_STATE_BADGE: Record<string, string> = {
  draft: "bg-slate-100 text-slate-700",
  negociation: "bg-amber-100 text-amber-700",
  validated: "bg-green-100 text-green-700",
  drop: "bg-red-100 text-red-700",
  sent: "bg-blue-100 text-blue-700",
  accepted: "bg-green-100 text-green-700",
  refused: "bg-red-100 text-red-700",
};

export default function UserDashboard() {
  const t = useTranslations("dashboard.user");
  const [details, setDetails] = useState<BackendProjectDetail[]>([]);
  const [recentProjects, setRecentProjects] = useState<BackendProject[]>([]);
  const [recentQuotes, setRecentQuotes] = useState<BackendQuote[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      const [projectsRes, quotesRes] = await Promise.all([
        listProjects("page_size=50&sort_by=created_at&sort_direction=desc"),
        listQuotes("page_size=10&sort_by=created_at&sort_direction=desc"),
      ]);

      if (quotesRes.ok && quotesRes.body.success) {
        setRecentQuotes((quotesRes.body.quotes ?? []) as BackendQuote[]);
      }

      if (projectsRes.ok && projectsRes.body.success) {
        const projects = (projectsRes.body.projects ?? []) as BackendProject[];
        setRecentProjects(projects.slice(0, 10));

        const detailResults = await Promise.all(
          projects.map((p) => getProjectDetail(p.project_id)),
        );
        setDetails(
          detailResults
            .filter((r) => r.ok && r.body.success)
            .map((r) => r.body as BackendProjectDetail),
        );
      }

      setLoading(false);
    }
    load();
  }, []);

  const stateCount: Record<string, number> = {};
  const scheduleCount: Record<string, number> = {};
  let totalHtCents = 0;
  let collectedHtCents = 0;

  for (const d of details) {
    totalHtCents += d.total_ht_cents ?? 0;
    collectedHtCents += d.collected_ht_cents ?? 0;
    for (const q of d.quotes ?? []) {
      stateCount[q.state] = (stateCount[q.state] ?? 0) + 1;
      for (const s of q.schedules ?? []) {
        scheduleCount[s.status] = (scheduleCount[s.status] ?? 0) + 1;
      }
    }
  }

  const statePieData = Object.entries(stateCount).map(([state, count]) => ({
    name: QUOTE_STATE_LABELS[state] ?? state,
    value: count,
    color: QUOTE_STATE_COLORS[state] ?? "#6b7280",
  }));

  const scheduleBarData = Object.entries(scheduleCount).map(([status, count]) => ({
    name: SCHEDULE_STATUS_LABELS[status] ?? status,
    count,
    color: SCHEDULE_STATUS_COLORS[status] ?? "#6b7280",
  }));

  const revenueData = [
    { name: "Total HT contractualisé", montant: totalHtCents / 100 },
    { name: "Encaissé", montant: collectedHtCents / 100 },
  ];

  const hasData = details.length > 0;

  return (
    <div className="space-y-8">
      {/* Graphiques agrégés */}
      <Card>
        <CardHeader>
          <CardTitle>{t("chartsTitle")}</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
              {Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-48 w-full rounded-lg" />
              ))}
            </div>
          ) : hasData ? (
            <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
              <PieChartCard
                title="Répartition des devis par état"
                data={statePieData}
                outerRadius={70}
                height={200}
              />
              <BarChartCard
                title="Répartition des échéanciers"
                data={scheduleBarData}
                dataKey="count"
                colorKey="color"
                barName="Nb"
                height={200}
                noDataMessage="Aucun échéancier."
              />
              <BarChartCard
                title="Chiffre d'affaires (HT)"
                data={revenueData}
                dataKey="montant"
                defaultColor="#3b82f6"
                barName="Montant HT"
                height={200}
                tickFormatter={(v) => `${(v / 1000).toFixed(0)}k`}
                tooltipFormatter={(v) => formatEuros(v * 100)}
              />
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">{t("noData")}</p>
          )}
        </CardContent>
      </Card>

      {/* Listings */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* 10 derniers projets */}
        <Card>
          <CardHeader>
            <CardTitle>{t("recentProjects")}</CardTitle>
          </CardHeader>
          <CardContent>
            {loading ? (
              <div className="space-y-2">
                {Array.from({ length: 5 }).map((_, i) => (
                  <Skeleton key={i} className="h-8 w-full rounded" />
                ))}
              </div>
            ) : recentProjects.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t("noProjects")}</p>
            ) : (
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b text-left text-muted-foreground">
                    <th className="pb-2 font-medium">{t("col.name")}</th>
                    <th className="pb-2 font-medium">{t("col.status")}</th>
                    <th className="pb-2 font-medium">{t("col.date")}</th>
                  </tr>
                </thead>
                <tbody>
                  {recentProjects.map((p) => (
                    <tr key={p.project_id} className="border-b last:border-0">
                      <td className="py-2">
                        <Link
                          href={`/projects/${p.project_id}`}
                          className="hover:underline"
                        >
                          {p.name}
                        </Link>
                      </td>
                      <td className="py-2 capitalize text-muted-foreground">
                        {p.status}
                      </td>
                      <td className="py-2 text-muted-foreground">
                        {new Date(p.created_at).toLocaleDateString("fr-FR")}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </CardContent>
        </Card>

        {/* 10 derniers devis */}
        <Card>
          <CardHeader>
            <CardTitle>{t("recentQuotes")}</CardTitle>
          </CardHeader>
          <CardContent>
            {loading ? (
              <div className="space-y-2">
                {Array.from({ length: 5 }).map((_, i) => (
                  <Skeleton key={i} className="h-8 w-full rounded" />
                ))}
              </div>
            ) : recentQuotes.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t("noQuotes")}</p>
            ) : (
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b text-left text-muted-foreground">
                    <th className="pb-2 font-medium">{t("col.name")}</th>
                    <th className="pb-2 font-medium">{t("col.state")}</th>
                    <th className="pb-2 font-medium">{t("col.total")}</th>
                  </tr>
                </thead>
                <tbody>
                  {recentQuotes.map((q) => (
                    <tr key={q.quote_id} className="border-b last:border-0">
                      <td className="py-2">
                        <Link
                          href={`/quote/${q.quote_id}`}
                          className="hover:underline"
                        >
                          {q.name}
                        </Link>
                      </td>
                      <td className="py-2">
                        <span
                          className={`rounded px-2 py-0.5 text-xs font-medium ${QUOTE_STATE_BADGE[q.state] ?? "bg-gray-100 text-gray-700"}`}
                        >
                          {QUOTE_STATE_LABELS[q.state] ?? q.state}
                        </span>
                      </td>
                      <td className="py-2 text-muted-foreground">
                        {q.total_ttc != null ? formatEuros(q.total_ttc) : "—"}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
