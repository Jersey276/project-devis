"use client";

import { useState, useEffect } from "react";
import { useTranslations } from "next-intl";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { apiFetch } from "@/lib/api";
import type { ActivityLog, ActivityLogDetail } from "./logs-dashboard";

type LogsTableProps = {
  logs: ActivityLog[];
};

const PREVIEW_LIMIT = 500;

function statusBadgeClass(status: number): string {
  if (status >= 500) return "text-red-600 font-semibold";
  if (status >= 400) return "text-orange-500 font-semibold";
  if (status >= 300) return "text-yellow-600";
  return "text-green-600";
}

function tryFormatJSON(raw: string): string {
  if (!raw) return "";
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    if (raw.includes("=")) {
      try {
        const params = Object.fromEntries(new URLSearchParams(raw));
        return JSON.stringify(params, null, 2);
      } catch {
        return raw;
      }
    }
    return raw;
  }
}

function BodyPanel({ label, value }: { label: string; value: string }) {
  const [expanded, setExpanded] = useState(false);
  if (!value) return null;
  const formatted = tryFormatJSON(value);
  const truncated = !expanded && formatted.length > PREVIEW_LIMIT;
  const displayed = truncated ? formatted.slice(0, PREVIEW_LIMIT) + "…" : formatted;
  return (
    <div className="space-y-1">
      <p className="text-xs font-medium text-muted-foreground">{label}</p>
      <pre className="overflow-auto rounded bg-muted px-3 py-2 font-mono text-xs whitespace-pre-wrap break-all">
        {displayed}
      </pre>
      {formatted.length > PREVIEW_LIMIT && (
        <button
          className="text-xs text-primary underline"
          onClick={() => setExpanded((v) => !v)}
        >
          {expanded ? "Réduire" : "Voir tout"}
        </button>
      )}
    </div>
  );
}

type DetailPanelProps = {
  detail: ActivityLogDetail | null;
  error: boolean;
  tTable: ReturnType<typeof useTranslations>;
};

function DetailPanel({ detail, error, tTable }: DetailPanelProps) {
  return (
    <TableRow>
      <TableCell colSpan={8} className="bg-muted/30 px-4 py-3">
        {!detail && !error && (
          <p className="text-xs text-muted-foreground">Chargement…</p>
        )}
        {error && (
          <p className="text-xs text-destructive">Erreur lors du chargement.</p>
        )}
        {detail && (
          <div className="space-y-3">
            <BodyPanel label={tTable("reqBody")} value={detail.req_body} />
            <BodyPanel label={tTable("respBody")} value={detail.resp_body} />
          </div>
        )}
      </TableCell>
    </TableRow>
  );
}

export default function LogsTable({ logs }: LogsTableProps) {
  const t = useTranslations("admin.logs.table");
  const [expandedId, setExpandedId] = useState<number | null>(null);
  const [cache, setCache] = useState<Record<number, ActivityLogDetail | "error">>({});

  useEffect(() => {
    if (expandedId === null || cache[expandedId] !== undefined) return;
    apiFetch(`/api/logs/${expandedId}`).then(({ ok, body }) => {
      const id = expandedId;
      setCache((prev) => ({
        ...prev,
        [id]: ok && body.success ? (body.log as ActivityLogDetail) : "error",
      }));
    });
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [expandedId]);

  const toggle = (id: number) =>
    setExpandedId((prev) => (prev === id ? null : id));

  return (
    <div className="overflow-x-auto rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-8" />
            <TableHead className="w-16 hidden sm:table-cell">{t("id")}</TableHead>
            <TableHead className="hidden md:table-cell">{t("userId")}</TableHead>
            <TableHead className="w-20">{t("method")}</TableHead>
            <TableHead>{t("url")}</TableHead>
            <TableHead className="w-28 text-right hidden sm:table-cell">{t("durationMs")}</TableHead>
            <TableHead className="w-20 text-center">{t("respStatus")}</TableHead>
            <TableHead className="w-40 hidden lg:table-cell">{t("createdAt")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {logs.map((log) => (
            <>
              <TableRow
                key={log.id}
                className="cursor-pointer hover:bg-muted/50"
                onClick={() => toggle(log.id)}
              >
                <TableCell className="text-center text-muted-foreground">
                  <span className="text-xs">{expandedId === log.id ? "▾" : "▸"}</span>
                </TableCell>
                <TableCell className="text-muted-foreground hidden sm:table-cell">{log.id}</TableCell>
                <TableCell className="max-w-30 truncate text-xs hidden md:table-cell" title={log.user_id}>
                  {log.user_id || "—"}
                </TableCell>
                <TableCell>
                  <span className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">
                    {log.method}
                  </span>
                </TableCell>
                <TableCell className="max-w-40 sm:max-w-64 truncate font-mono text-xs" title={log.url}>
                  {log.url}
                </TableCell>
                <TableCell className="text-right font-mono text-xs hidden sm:table-cell">
                  {log.duration_ms} ms
                </TableCell>
                <TableCell className={`text-center font-mono text-sm ${statusBadgeClass(log.resp_status)}`}>
                  {log.resp_status}
                </TableCell>
                <TableCell className="text-xs text-muted-foreground hidden lg:table-cell">
                  {new Date(log.created_at).toLocaleString("fr-FR")}
                </TableCell>
              </TableRow>

              {expandedId === log.id && (
                <DetailPanel
                  key={`${log.id}-detail`}
                  detail={cache[log.id] instanceof Object ? (cache[log.id] as ActivityLogDetail) : null}
                  error={cache[log.id] === "error"}
                  tTable={t}
                />
              )}
            </>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
