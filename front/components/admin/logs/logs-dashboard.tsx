"use client";

import { useEffect, useState, useCallback } from "react";
import { useSearchParams, useRouter, usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { apiFetch } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";
import LogsStatsChart from "./logs-stats-chart";
import LogsTable from "./logs-table";
import LogsFilters, { type LogFilters } from "./logs-filters";
import ExportButton from "./export-button";

export type ActivityLog = {
  id: number;
  user_id: string;
  method: string;
  url: string;
  duration_ms: number;
  resp_status: number;
  created_at: string;
};

export type ActivityLogDetail = {
  id: number;
  user_id: string;
  method: string;
  url: string;
  duration_ms: number;
  req_body: string;
  resp_body: string;
  resp_status: number;
  created_at: string;
};

const PAGE_SIZE = 50;

export default function LogsDashboard() {
  const t = useTranslations("admin.logs");
  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();

  const page = Number(searchParams.get("page") ?? "1");
  const filters: LogFilters = {
    search: searchParams.get("search") ?? "",
    resp_statuses: searchParams.get("resp_statuses")
      ? searchParams.get("resp_statuses")!.split(",").map(Number)
      : [],
    date_from: searchParams.get("date_from") ?? "",
    date_to: searchParams.get("date_to") ?? "",
  };

  const [logs, setLogs] = useState<ActivityLog[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);

  function pushParams(newFilters: LogFilters, newPage: number) {
    const p = new URLSearchParams();
    if (newPage > 1) p.set("page", String(newPage));
    if (newFilters.search) p.set("search", newFilters.search);
    if (newFilters.resp_statuses.length > 0)
      p.set("resp_statuses", newFilters.resp_statuses.join(","));
    if (newFilters.date_from) p.set("date_from", newFilters.date_from);
    if (newFilters.date_to) p.set("date_to", newFilters.date_to);
    router.push(`${pathname}?${p.toString()}`);
  }

  const handleFiltersChange = (newFilters: LogFilters) => pushParams(newFilters, 1);

  const fetchLogs = useCallback(async () => {
    setLoading(true);
    const params = new URLSearchParams({
      page: String(page),
      page_size: String(PAGE_SIZE),
    });
    if (filters.search) {
      params.set("url_contains", filters.search);
      params.set("user_id", filters.search);
    }
    if (filters.resp_statuses.length > 0) params.set("resp_statuses", filters.resp_statuses.join(","));
    if (filters.date_from) params.set("date_from", filters.date_from);
    if (filters.date_to) params.set("date_to", filters.date_to);

    const { ok, body } = await apiFetch(`/api/logs?${params.toString()}`);
    if (ok && body.success) {
      setLogs((body.logs ?? []) as ActivityLog[]);
      setTotal((body.total ?? 0) as number);
    }
    setLoading(false);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <div className="space-y-6">
      <LogsStatsChart />

      <div className="flex flex-wrap items-center justify-between gap-2">
        <LogsFilters filters={filters} onChange={handleFiltersChange} />
        <ExportButton filters={filters} />
      </div>

      {loading ? (
        <div className="space-y-2">
          {Array.from({ length: 8 }).map((_, i) => <Skeleton key={i} className="h-10 w-full" />)}
        </div>
      ) : logs.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("noData")}</p>
      ) : (
        <>
          <LogsTable logs={logs} />
          <div className="flex flex-wrap items-center justify-between gap-2 text-sm text-muted-foreground">
            <span>
              {total} résultat{total > 1 ? "s" : ""}
            </span>
            <div className="flex gap-2">
              <button
                className="rounded border px-3 py-1 disabled:opacity-40"
                disabled={page <= 1}
                onClick={() => pushParams(filters, page - 1)}
              >
                ←
              </button>
              <span>
                {page} / {totalPages}
              </span>
              <button
                className="rounded border px-3 py-1 disabled:opacity-40"
                disabled={page >= totalPages}
                onClick={() => pushParams(filters, page + 1)}
              >
                →
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
