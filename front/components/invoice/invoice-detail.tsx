"use client";

import { useCallback, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import InvoiceStatusBadge from "@/components/invoice/invoice-status-badge";
import {
  cancelInvoice,
  getInvoice,
  markInvoicePaid,
  readInvoiceFromBody,
} from "@/lib/services/invoices";
import { exportInvoicePdf } from "@/lib/services/export";
import { formatEurosFromCents } from "@/lib/utils";
import type { BackendInvoiceDetails, BackendInvoiceParty } from "@/types/backend";

function partyLines(p: BackendInvoiceParty | undefined): string[] {
  if (!p) return [];
  const lines: string[] = [];
  const title =
    p.company || `${p.first_name} ${p.last_name}`.trim();
  if (title) lines.push(title);
  if (p.street) lines.push(p.street);
  if (p.additional_street) lines.push(p.additional_street);
  const city = `${p.zip_code} ${p.city}`.trim();
  if (city) lines.push(city);
  if (p.email) lines.push(p.email);
  if (p.phone) lines.push(p.phone);
  if (p.siren) lines.push(`SIREN : ${p.siren}`);
  if (p.vat) lines.push(`TVA : ${p.vat}`);
  return lines;
}

export default function InvoiceDetail({ invoiceId }: { invoiceId: string }) {
  const t = useTranslations("invoice.detail");
  const [invoice, setInvoice] = useState<BackendInvoiceDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const load = useCallback(async () => {
    const { ok, body } = await getInvoice(invoiceId);
    setLoading(false);
    if (!ok || !body.success) {
      setError((body.message as string) ?? t("loadError"));
      return;
    }
    setError(null);
    setInvoice(readInvoiceFromBody(body));
  }, [invoiceId, t]);

  useEffect(() => {
    let cancelled = false;
    getInvoice(invoiceId).then(({ ok, body }) => {
      if (cancelled) return;
      setLoading(false);
      if (!ok || !body.success) {
        setError((body.message as string) ?? t("loadError"));
        return;
      }
      setError(null);
      setInvoice(readInvoiceFromBody(body));
    });
    return () => {
      cancelled = true;
    };
  }, [invoiceId, t]);

  async function onDownload() {
    try {
      await exportInvoicePdf(invoiceId);
    } catch {
      setError(t("pdfError"));
    }
  }

  async function onMarkPaid() {
    setBusy(true);
    const { ok, body } = await markInvoicePaid(invoiceId);
    setBusy(false);
    if (!ok || !body.success) setError((body.message as string) ?? t("actionError"));
    else void load();
  }

  async function onCancel() {
    setBusy(true);
    const { ok, body } = await cancelInvoice(invoiceId);
    setBusy(false);
    if (!ok || !body.success) setError((body.message as string) ?? t("actionError"));
    else void load();
  }

  if (loading) return <p>{t("loading")}</p>;
  if (error && !invoice) return <p className="text-destructive">{error}</p>;
  if (!invoice) return <p className="text-destructive">{t("notFound")}</p>;

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between gap-4">
        <CardTitle className="flex items-center gap-3">
          {invoice.invoice_number
            ? t("titleNumbered", { number: invoice.invoice_number })
            : t("titleDraft")}
          <InvoiceStatusBadge status={invoice.status} />
        </CardTitle>
        <div className="flex items-center gap-2">
          <Button type="button" variant="outline" onClick={onDownload}>
            {t("downloadPdf")}
          </Button>
          {invoice.status === "ISSUED" ? (
            <Button type="button" variant="outline" onClick={onMarkPaid} disabled={busy}>
              {t("markPaid")}
            </Button>
          ) : null}
          {invoice.status === "ISSUED" || invoice.status === "PAID" ? (
            <Button type="button" variant="destructive" onClick={onCancel} disabled={busy}>
              {t("cancel")}
            </Button>
          ) : null}
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {error ? <p className="text-sm text-destructive">{error}</p> : null}

        <div className="grid grid-cols-2 gap-6 text-sm">
          <section>
            <h3 className="mb-1 font-semibold">{t("issuer")}</h3>
            {partyLines(invoice.issuer).map((line, i) => (
              <div key={i}>{line}</div>
            ))}
          </section>
          <section>
            <h3 className="mb-1 font-semibold">{t("client")}</h3>
            {partyLines(invoice.client).map((line, i) => (
              <div key={i}>{line}</div>
            ))}
          </section>
        </div>

        <div className="grid grid-cols-3 gap-4 text-sm text-muted-foreground">
          <div>{t("saleDate")}: {invoice.sale_date || "—"}</div>
          <div>{t("dueDate")}: {invoice.due_date || "—"}</div>
          <div>{t("issuedAt")}: {invoice.issued_at || "—"}</div>
        </div>

        <table className="w-full border-collapse text-sm">
          <thead>
            <tr className="border-b text-left">
              <th className="py-1">{t("lines.name")}</th>
              <th className="py-1 text-center">{t("lines.quantity")}</th>
              <th className="py-1 text-center">{t("lines.unit")}</th>
              <th className="py-1 text-right">{t("lines.unitPrice")}</th>
              {!invoice.vat_exempt ? (
                <th className="py-1 text-center">{t("lines.vat")}</th>
              ) : null}
              <th className="py-1 text-right">{t("lines.totalHt")}</th>
            </tr>
          </thead>
          <tbody>
            {invoice.lines.map((l, i) => (
              <tr key={i} className="border-b">
                <td className="py-1">{l.name}</td>
                <td className="py-1 text-center">{l.quantity}</td>
                <td className="py-1 text-center">{l.unit}</td>
                <td className="py-1 text-right tabular-nums">
                  {formatEurosFromCents(l.unit_price_cents)}
                </td>
                {!invoice.vat_exempt ? (
                  <td className="py-1 text-center">{l.tax_rate} %</td>
                ) : null}
                <td className="py-1 text-right tabular-nums">
                  {formatEurosFromCents(l.line_ht_cents)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        <div className="ml-auto w-64 space-y-1 text-sm">
          <div className="flex justify-between">
            <span>{t("totals.ht")}</span>
            <span className="tabular-nums">
              {formatEurosFromCents(invoice.total_ht_cents)}
            </span>
          </div>
          {!invoice.vat_exempt ? (
            <div className="flex justify-between">
              <span>{t("totals.vat")}</span>
              <span className="tabular-nums">
                {formatEurosFromCents(invoice.total_vat_cents)}
              </span>
            </div>
          ) : null}
          <div className="flex justify-between border-t pt-1 font-semibold">
            <span>{t("totals.ttc")}</span>
            <span className="tabular-nums">
              {formatEurosFromCents(invoice.total_ttc_cents)}
            </span>
          </div>
        </div>

        {invoice.vat_exempt ? (
          <p className="text-xs text-muted-foreground">{t("vatExemptNotice")}</p>
        ) : null}
      </CardContent>
    </Card>
  );
}
