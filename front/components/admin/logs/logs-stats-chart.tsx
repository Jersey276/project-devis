"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { apiFetch } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";
import LineChartCard, { type LineSeriesConfig } from "@/components/charts/line-chart-card";

type RawStat = {
  date: string;
  resp_status: number;
  count: number;
};

type ChartPoint = {
  date: string;
  [statusGroup: string]: number | string;
};

const STATUS_GROUPS: Record<string, { label: string; color: string }> = {
  "2xx": { label: "2xx", color: "#22c55e" },
  "3xx": { label: "3xx", color: "#eab308" },
  "4xx": { label: "4xx", color: "#f97316" },
  "5xx": { label: "5xx", color: "#ef4444" },
};

function groupKey(status: number): string {
  if (status >= 500) return "5xx";
  if (status >= 400) return "4xx";
  if (status >= 300) return "3xx";
  if (status >= 200) return "2xx";
  return "other";
}

function pivot(raw: RawStat[]): ChartPoint[] {
  const byDate = new Map<string, ChartPoint>();
  for (const entry of raw) {
    if (!byDate.has(entry.date)) {
      byDate.set(entry.date, { date: entry.date });
    }
    const point = byDate.get(entry.date)!;
    const key = groupKey(entry.resp_status);
    point[key] = ((point[key] as number | undefined) ?? 0) + entry.count;
  }
  return Array.from(byDate.values()).sort((a, b) => (a.date < b.date ? -1 : 1));
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr + "T00:00:00Z");
  return d.toLocaleDateString("fr-FR", { day: "2-digit", month: "2-digit", timeZone: "UTC" });
}

export default function LogsStatsChart() {
  const t = useTranslations("admin.logs");
  const [data, setData] = useState<ChartPoint[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    apiFetch("/api/logs/stats").then(({ ok, body }) => {
      if (ok && body.success) {
        setData(pivot((body.stats ?? []) as RawStat[]));
      }
      setLoading(false);
    });
  }, []);

  const presentGroups = Array.from(
    new Set(data.flatMap((p) => Object.keys(p).filter((k) => k !== "date")))
  ).sort();

  const lines: LineSeriesConfig[] = presentGroups.map((group) => ({
    key: group,
    color: STATUS_GROUPS[group]?.color ?? "#94a3b8",
    label: `HTTP ${group}`,
  }));

  if (loading) {
    return <Skeleton className="h-48 w-full rounded-lg" />;
  }

  if (data.length === 0) {
    return null;
  }

  return (
    <LineChartCard
      title={t("statsTitle")}
      data={data}
      lines={lines}
      xAxisKey="date"
      height={240}
      xTickFormatter={formatDate}
      tooltipLabelFormatter={formatDate}
      tooltipFormatter={(value, name) => [value, name]}
      vertical={false}
    />
  );
}
