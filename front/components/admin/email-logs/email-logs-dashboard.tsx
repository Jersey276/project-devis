"use client";

import { useEffect, useState, useCallback } from "react";
import { useSearchParams, useRouter, usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { apiFetch } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";
import EmailLogsFilters, { type EmailLogFilters } from "./email-logs-filters";
import EmailLogsTable from "./email-logs-table";

export type EmailLog = {
  id: number;
  to_email: string;
  type: string;
  reference_name: string | null;
  status: string;
  created_at: string;
  opened?: boolean;
  clicked?: boolean;
};

const PAGE_SIZE = 20;

export default function EmailLogsDashboard() {
  const t = useTranslations("admin.emailLogs");
  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();

  const page = Number(searchParams.get("page") ?? "1");
  const filters: EmailLogFilters = {
    search: searchParams.get("search") ?? "",
    statuses: searchParams.get("statuses") ? searchParams.get("statuses")!.split(",") : [],
    types: searchParams.get("types") ? searchParams.get("types")!.split(",") : [],
    date_from: searchParams.get("date_from") ?? "",
    date_to: searchParams.get("date_to") ?? "",
  };

  const [logs, setLogs] = useState<EmailLog[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);

  function pushParams(newFilters: EmailLogFilters, newPage: number) {
    const p = new URLSearchParams();
    if (newPage > 1) p.set("page", String(newPage));
    if (newFilters.search) p.set("search", newFilters.search);
    if (newFilters.statuses.length > 0) p.set("statuses", newFilters.statuses.join(","));
    if (newFilters.types.length > 0) p.set("types", newFilters.types.join(","));
    if (newFilters.date_from) p.set("date_from", newFilters.date_from);
    if (newFilters.date_to) p.set("date_to", newFilters.date_to);
    router.push(`${pathname}?${p.toString()}`);
  }

  const handleFiltersChange = (newFilters: EmailLogFilters) => pushParams(newFilters, 1);

  const fetchLogs = useCallback(async () => {
    setLoading(true);
    const params = new URLSearchParams({
      limit: String(PAGE_SIZE),
      offset: String((page - 1) * PAGE_SIZE),
    });
    if (filters.search) params.set("search", filters.search);
    if (filters.statuses.length > 0) params.set("statuses", filters.statuses.join(","));
    if (filters.types.length > 0) params.set("types", filters.types.join(","));
    if (filters.date_from) params.set("date_from", filters.date_from);
    if (filters.date_to) params.set("date_to", filters.date_to);

    const { ok, body } = await apiFetch(`/api/email-logs?${params.toString()}`);
    if (ok && body.success) {
      setLogs((body.logs ?? []) as EmailLog[]);
      setTotal((body.total ?? 0) as number);
    }
    setLoading(false);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams]);

  // eslint-disable-next-line react-hooks/set-state-in-effect
  useEffect(() => { void fetchLogs(); }, [fetchLogs]);

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <EmailLogsFilters filters={filters} onChange={handleFiltersChange} />
      </div>

      {loading ? (
        <div className="space-y-2">
          {Array.from({ length: 8 }).map((_, i) => (
            <Skeleton key={i} className="h-10 w-full" />
          ))}
        </div>
      ) : logs.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("noData")}</p>
      ) : (
        <>
          <EmailLogsTable logs={logs} />
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
