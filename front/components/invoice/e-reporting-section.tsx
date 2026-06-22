"use client";

import { useCallback, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import InvoiceLifecycleBadge from "@/components/invoice/invoice-lifecycle-badge";
import {
  listInvoiceReports,
  readInvoiceReportsFromBody,
  submitInvoiceReport,
} from "@/lib/services/invoices";
import { formatEurosFromCents } from "@/lib/utils";
import type {
  BackendInvoiceReportSummary,
  BackendReportKind,
} from "@/types/backend";

const KINDS: BackendReportKind[] = ["TRANSACTION", "CROSS_BORDER_B2C"];

// E-reporting (B5/C5): transmit and track the monthly out-of-scope-of-e-invoicing
// aggregates (domestic B2C and intra-EU distance sales). The platform is a no-op
// until a PA is contracted, so submitting just records the period locally for now.
export default function EReportingSection() {
  const t = useTranslations("invoice.reporting");
  const now = new Date();
  const [kind, setKind] = useState<BackendReportKind>("TRANSACTION");
  const [year, setYear] = useState(now.getFullYear());
  // Default to the previous month: the period being reported is usually closed.
  const [month, setMonth] = useState(now.getMonth() === 0 ? 12 : now.getMonth());
  const [reports, setReports] = useState<BackendInvoiceReportSummary[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(() => {
    listInvoiceReports().then(({ ok, body }) => {
      if (!ok || !body.success) return;
      setReports(readInvoiceReportsFromBody(body));
    });
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  async function onSubmit() {
    setBusy(true);
    setError(null);
    const { ok, body } = await submitInvoiceReport(kind, year, month);
    setBusy(false);
    if (!ok || !body.success) {
      setError((body.message as string) ?? t("submitError"));
      return;
    }
    refresh();
  }

  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-lg font-semibold">{t("title")}</h2>
        <p className="text-sm text-muted-foreground">{t("description")}</p>
      </div>

      <div className="flex flex-wrap items-end gap-3">
        <Select value={kind} onValueChange={(v) => setKind(v as BackendReportKind)}>
          <SelectTrigger className="w-64">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {KINDS.map((k) => (
              <SelectItem key={k} value={k}>
                {t(`kind.${k}`)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select
          value={String(month)}
          onValueChange={(v) => setMonth(Number(v))}
        >
          <SelectTrigger className="w-40">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {Array.from({ length: 12 }, (_, i) => i + 1).map((m) => (
              <SelectItem key={m} value={String(m)}>
                {t(`month.${m}`)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select value={String(year)} onValueChange={(v) => setYear(Number(v))}>
          <SelectTrigger className="w-28">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {[now.getFullYear(), now.getFullYear() - 1].map((y) => (
              <SelectItem key={y} value={String(y)}>
                {y}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Button type="button" onClick={onSubmit} disabled={busy}>
          {t("submit")}
        </Button>
      </div>

      {error ? <p className="text-sm text-destructive">{error}</p> : null}

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t("col.kind")}</TableHead>
            <TableHead>{t("col.period")}</TableHead>
            <TableHead className="text-right">{t("col.totalHt")}</TableHead>
            <TableHead className="text-right">{t("col.totalVat")}</TableHead>
            <TableHead>{t("col.status")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {reports.length === 0 ? (
            <TableRow>
              <TableCell colSpan={5} className="text-muted-foreground">
                {t("empty")}
              </TableCell>
            </TableRow>
          ) : (
            reports.map((r) => (
              <TableRow key={`${r.kind}-${r.year}-${r.month}`}>
                <TableCell>{t(`kind.${r.kind}`)}</TableCell>
                <TableCell>
                  {t(`month.${r.month}`)} {r.year}
                </TableCell>
                <TableCell className="text-right">
                  {formatEurosFromCents(r.total_ht_cents)}
                </TableCell>
                <TableCell className="text-right">
                  {formatEurosFromCents(r.total_vat_cents)}
                </TableCell>
                <TableCell>
                  <InvoiceLifecycleBadge status={r.status} />
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </section>
  );
}
