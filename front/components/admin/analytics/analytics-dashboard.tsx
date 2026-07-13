"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { getAdminStats } from "@/lib/services/subscriptions";
import { Skeleton } from "@/components/ui/skeleton";
import MetricCard from "@/components/charts/metric-card";
import PieChartCard from "@/components/charts/pie-chart-card";
import LineChartCard from "@/components/charts/line-chart-card";
import type { AdminStats } from "@/types/backend";

const TIER_COLORS: Record<string, string> = {
  free: "#94a3b8",
  pro: "#3b82f6",
  enterprise: "#7c3aed",
};

const TIER_LABELS: Record<string, string> = {
  free: "Gratuit",
  pro: "Pro",
  enterprise: "Enterprise",
};

function formatEuros(cents: number): string {
  return new Intl.NumberFormat("fr-FR", {
    style: "currency",
    currency: "EUR",
    maximumFractionDigits: 0,
  }).format(cents / 100);
}

function formatMonth(month: string): string {
  const d = new Date(month.length === 7 ? `${month}-01T00:00:00Z` : `${month}T00:00:00Z`);
  return d.toLocaleDateString("fr-FR", { month: "2-digit", year: "numeric", timeZone: "UTC" });
}

export default function AnalyticsDashboard() {
  const t = useTranslations("admin.analytics");
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    getAdminStats().then(({ ok, body }) => {
      if (cancelled) return;
      if (ok) {
        setStats(body as unknown as AdminStats);
      }
      setLoading(false);
    });
    return () => {
      cancelled = true;
    };
  }, []);

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-24 w-full rounded-lg" />)}
        </div>
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
          <Skeleton className="h-64 w-full rounded-lg" />
          <Skeleton className="h-64 w-full rounded-lg" />
        </div>
      </div>
    );
  }

  if (!stats) {
    return <p className="text-sm text-muted-foreground">{t("noData")}</p>;
  }

  const lastMonthRevenue =
    stats.monthly_revenue.length > 0
      ? stats.monthly_revenue[stats.monthly_revenue.length - 1].revenue_cents
      : 0;

  const pieData = stats.plan_distribution.map((e) => ({
    name: TIER_LABELS[e.tier] ?? e.tier,
    value: e.count,
    color: TIER_COLORS[e.tier] ?? "#64748b",
  }));

  const lineData = stats.monthly_revenue.map((e) => ({
    month: e.month,
    revenue: e.revenue_cents / 100,
  }));

  return (
    <div className="grid gap-8">
      <div className="grid gap-4 sm:grid-cols-3">
        <MetricCard label={t("totalActive")} value={stats.total_active_subscriptions} />
        <MetricCard label={t("totalRevenue")} value={formatEuros(stats.total_revenue_cents)} />
        <MetricCard label={t("monthRevenue")} value={formatEuros(lastMonthRevenue)} />
      </div>

      <div className="grid gap-8 lg:grid-cols-2">
        <PieChartCard
          title={t("planDistribution")}
          data={pieData}
          innerRadius={60}
          outerRadius={100}
          height={280}
          showLegend
          tooltipSuffix=" abonnements"
        />
        <LineChartCard
          title={t("monthlyRevenue")}
          data={lineData}
          lines={[{ key: "revenue", color: "#3b82f6", label: "Revenu" }]}
          xAxisKey="month"
          height={280}
          xTickFormatter={formatMonth}
          tooltipLabelFormatter={formatMonth}
          yTickFormatter={(v) => `${v}€`}
          tooltipFormatter={(v) => [`${v}€`, "Revenu"]}
          vertical={false}
        />
      </div>
    </div>
  );
}
