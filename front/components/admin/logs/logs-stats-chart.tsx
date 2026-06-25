"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { apiFetch } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";

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

  return Array.from(byDate.values()).sort((a, b) =>
    a.date < b.date ? -1 : 1,
  );
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

  if (loading) {
    return <Skeleton className="h-48 w-full rounded-lg" />;
  }

  if (data.length === 0) {
    return null;
  }

  return (
    <div className="space-y-2">
      <p className="text-sm font-medium">{t("statsTitle")}</p>
      <ResponsiveContainer width="100%" height={220}>
        <LineChart data={data} margin={{ top: 4, right: 8, left: 0, bottom: 4 }}>
          <CartesianGrid strokeDasharray="3 3" vertical={false} />
          <XAxis
            dataKey="date"
            tickFormatter={formatDate}
            tick={{ fontSize: 11 }}
            tickLine={false}
            axisLine={false}
          />
          <YAxis
            allowDecimals={false}
            tick={{ fontSize: 11 }}
            tickLine={false}
            axisLine={false}
            width={32}
          />
          <Tooltip
            formatter={(value, name) => [value, `HTTP ${name}`]}
            labelFormatter={(label) => formatDate(label as string)}
          />
          {presentGroups.map((group) => (
            <Line
              key={group}
              type="monotone"
              dataKey={group}
              stroke={STATUS_GROUPS[group]?.color ?? "#94a3b8"}
              strokeWidth={2}
              dot={false}
              activeDot={{ r: 4 }}
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
