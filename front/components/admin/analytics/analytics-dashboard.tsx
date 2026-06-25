"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  PieChart,
  Pie,
  Cell,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";
import { getAdminStats } from "@/lib/services/subscriptions";
import { Skeleton } from "@/components/ui/skeleton";
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

type MetricCardProps = {
  label: string;
  value: string | number;
};

function MetricCard({ label, value }: MetricCardProps) {
  return (
    <div className="rounded-lg border bg-card p-4 text-card-foreground shadow-sm">
      <p className="text-sm text-muted-foreground">{label}</p>
      <p className="mt-1 text-2xl font-bold">{value}</p>
    </div>
  );
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
    tier: e.tier,
  }));

  const lineData = stats.monthly_revenue.map((e) => ({
    month: e.month,
    revenue: e.revenue_cents / 100,
  }));

  return (
    <div className="grid gap-8">
      <div className="grid gap-4 sm:grid-cols-3">
        <MetricCard
          label={t("totalActive")}
          value={stats.total_active_subscriptions}
        />
        <MetricCard
          label={t("totalRevenue")}
          value={formatEuros(stats.total_revenue_cents)}
        />
        <MetricCard
          label={t("monthRevenue")}
          value={formatEuros(lastMonthRevenue)}
        />
      </div>

      <div className="grid gap-8 lg:grid-cols-2">
        <div>
          <h3 className="mb-4 text-sm font-medium">{t("planDistribution")}</h3>
          <ResponsiveContainer width="100%" height={260}>
            <PieChart>
              <Pie
                data={pieData}
                cx="50%"
                cy="50%"
                innerRadius={60}
                outerRadius={100}
                dataKey="value"
                label={({ name, value }) => `${name} (${value})`}
                labelLine={false}
              >
                {pieData.map((entry) => (
                  <Cell
                    key={entry.tier}
                    fill={TIER_COLORS[entry.tier] ?? "#64748b"}
                  />
                ))}
              </Pie>
              <Tooltip formatter={(v) => [v, "abonnements"]} />
              <Legend />
            </PieChart>
          </ResponsiveContainer>
        </div>

        <div>
          <h3 className="mb-4 text-sm font-medium">{t("monthlyRevenue")}</h3>
          <ResponsiveContainer width="100%" height={260}>
            <LineChart data={lineData}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis dataKey="month" tick={{ fontSize: 12 }} tickLine={false} />
              <YAxis
                tick={{ fontSize: 12 }}
                tickLine={false}
                tickFormatter={(v: number) => `${v}€`}
              />
              <Tooltip formatter={(v) => [`${v}€`, "Revenu"]} />
              <Line
                type="monotone"
                dataKey="revenue"
                stroke="#3b82f6"
                strokeWidth={2}
                dot={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  );
}
