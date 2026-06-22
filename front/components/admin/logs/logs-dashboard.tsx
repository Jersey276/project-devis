"use client";

import { useEffect, useState, useCallback } from "react";
import { useTranslations } from "next-intl";
import { apiFetch } from "@/lib/api";
import LogsStatsChart from "./logs-stats-chart";
import LogsTable from "./logs-table";
import LogsFilters, { type LogFilters, emptyLogFilters } from "./logs-filters";
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

  const [logs, setLogs] = useState<ActivityLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [filters, setFilters] = useState<LogFilters>(emptyLogFilters);

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
  }, [page, filters]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  const handleFiltersChange = (newFilters: LogFilters) => {
    setPage(1);
    setFilters(newFilters);
  };

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <div className="space-y-6">
      <LogsStatsChart />

      <div className="flex items-center justify-between gap-4">
        <LogsFilters filters={filters} onChange={handleFiltersChange} />
        <ExportButton filters={filters} />
      </div>

      {loading ? (
        <p className="text-sm text-muted-foreground">{t("loading")}</p>
      ) : logs.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("noData")}</p>
      ) : (
        <>
          <LogsTable logs={logs} />
          <div className="flex items-center justify-between text-sm text-muted-foreground">
            <span>
              {total} résultat{total > 1 ? "s" : ""}
            </span>
            <div className="flex gap-2">
              <button
                className="rounded border px-3 py-1 disabled:opacity-40"
                disabled={page <= 1}
                onClick={() => setPage((p) => p - 1)}
              >
                ←
              </button>
              <span>
                {page} / {totalPages}
              </span>
              <button
                className="rounded border px-3 py-1 disabled:opacity-40"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
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
