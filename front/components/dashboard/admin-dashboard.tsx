"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Skeleton } from "@/components/ui/skeleton";
import LogsStatsChart from "@/components/admin/logs/logs-stats-chart";
import PieChartCard from "@/components/charts/pie-chart-card";
import LineChartCard, { type LineSeriesConfig } from "@/components/charts/line-chart-card";
import { listAdminUsers } from "@/lib/services/admin-users";
import { type AdminUserAccount } from "@/components/admin/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const NOW = new Date();

function daysSince(dateStr: string | null): number | null {
  if (!dateStr) return null;
  return Math.floor((NOW.getTime() - new Date(dateStr).getTime()) / 86_400_000);
}

type LoginRange = "last7" | "last30" | "last90" | "older" | "never";

function loginRange(lastLoginAt: string | null): LoginRange {
  const days = daysSince(lastLoginAt);
  if (days === null) return "never";
  if (days <= 7) return "last7";
  if (days <= 30) return "last30";
  if (days <= 90) return "last90";
  return "older";
}

const LOGIN_RANGE_COLORS: Record<LoginRange, string> = {
  last7: "#22c55e",
  last30: "#3b82f6",
  last90: "#f59e0b",
  older: "#ef4444",
  never: "#94a3b8",
};

function buildLoginPieData(
  users: AdminUserAccount[],
  labels: Record<string, string>,
) {
  const counts: Partial<Record<LoginRange, number>> = {};
  for (const u of users) {
    const r = loginRange(u.last_login_at);
    counts[r] = (counts[r] ?? 0) + 1;
  }
  const order: LoginRange[] = ["last7", "last30", "last90", "older", "never"];
  return order
    .filter((r) => (counts[r] ?? 0) > 0)
    .map((r) => ({
      name: labels[r] ?? r,
      value: counts[r]!,
      color: LOGIN_RANGE_COLORS[r],
    }));
}

function buildRegistrationLineData(users: AdminUserAccount[]) {
  const byMonth: Record<string, number> = {};
  for (const u of users) {
    if (!u.created_at) continue;
    const month = u.created_at.slice(0, 7);
    byMonth[month] = (byMonth[month] ?? 0) + 1;
  }
  return Object.entries(byMonth)
    .sort(([a], [b]) => (a < b ? -1 : 1))
    .map(([month, count]) => ({ month, count }));
}

const registrationLines: LineSeriesConfig[] = [
  { key: "count", color: "#7c3aed", label: "Inscriptions" },
];

export default function AdminDashboard() {
  const t = useTranslations("dashboard.admin");
  const [users, setUsers] = useState<AdminUserAccount[]>([]);
  const [loadingUsers, setLoadingUsers] = useState(true);

  useEffect(() => {
    listAdminUsers("page_size=500").then(({ ok, body }) => {
      if (ok && body.success) {
        setUsers((body.users ?? []) as AdminUserAccount[]);
      }
      setLoadingUsers(false);
    });
  }, []);

  const rangeLabels: Record<string, string> = {
    last7: t("ranges.last7"),
    last30: t("ranges.last30"),
    last90: t("ranges.last90"),
    older: t("ranges.older"),
    never: t("ranges.never"),
  };

  const loginPieData = buildLoginPieData(users, rangeLabels);
  const registrationLineData = buildRegistrationLineData(users);

  return (
    <div className="space-y-8">
      {/* Connexions & inscriptions */}
      <Card>
        <CardHeader>
          <CardTitle>{t("usersTitle")}</CardTitle>
        </CardHeader>
        <CardContent>
          {loadingUsers ? (
            <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
              <Skeleton className="h-64 w-full rounded-lg" />
              <Skeleton className="h-64 w-full rounded-lg" />
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
              <PieChartCard
                title={t("connectionBreakdown")}
                data={loginPieData}
                outerRadius={90}
                height={260}
                showLegend
              />
              <LineChartCard
                title={t("registrationTrend")}
                data={registrationLineData}
                lines={registrationLines}
                xAxisKey="month"
                height={260}
                tooltipFormatter={(v) => [v, "Inscriptions"]}
                vertical={false}
              />
            </div>
          )}
        </CardContent>
      </Card>

      {/* Activité HTTP */}
      <Card>
        <CardHeader>
          <CardTitle>{t("httpActivity")}</CardTitle>
        </CardHeader>
        <CardContent>
          <LogsStatsChart />
        </CardContent>
      </Card>
    </div>
  );
}
