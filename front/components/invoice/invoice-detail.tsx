"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import InvoiceStatusBadge from "@/components/invoice/invoice-status-badge";
import InvoiceLifecycleBadge from "@/components/invoice/invoice-lifecycle-badge";
import { Skeleton } from "@/components/ui/skeleton";
import AdvanceLifecycleDialog from "@/components/invoice/advance-lifecycle-dialog";
import LifecycleTimeline from "@/components/invoice/lifecycle-timeline";
import CreateCreditNoteDialog from "@/components/invoice/create-credit-note-dialog";
import LinkedCreditNotes from "@/components/invoice/linked-credit-notes";
import { allowedNextLifecycleStatuses } from "@/lib/invoice-lifecycle";
import { useMode } from "@/lib/mode-context";
import {
  deleteDraftInvoice,
  depositInvoice,
  getInvoice,
  markInvoicePaid,
  readInvoiceFromBody,
} from "@/lib/services/invoices";
import { exportInvoiceFacturx, exportInvoicePdf } from "@/lib/services/export";
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
  if (p.siret) lines.push(`SIRET : ${p.siret}`);
  else if (p.siren) lines.push(`SIREN : ${p.siren}`);
  if (p.vat) lines.push(`TVA : ${p.vat}`);
  return lines;
}

export default function InvoiceDetail({ invoiceId }: { invoiceId: string }) {
  const t = useTranslations("invoice.detail");
  const tLifecycle = useTranslations("invoice.lifecycle");
  const { isCustomer } = useMode();
  const router = useRouter();
  const [invoice, setInvoice] = useState<BackendInvoiceDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [creditDialogOpen, setCreditDialogOpen] = useState(false);
  const [creditRefresh, setCreditRefresh] = useState(0);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [lifecycleDialogOpen, setLifecycleDialogOpen] = useState(false);
  const [lifecycleRefresh, setLifecycleRefresh] = useState(0);

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

  async function onDownloadFacturx() {
    try {
      await exportInvoiceFacturx(invoiceId);
    } catch {
      setError(t("facturxError"));
    }
  }

  async function onMarkPaid() {
    setBusy(true);
    const { ok, body } = await markInvoicePaid(invoiceId);
    setBusy(false);
    if (!ok || !body.success) setError((body.message as string) ?? t("actionError"));
    else void load();
  }

  async function onDeposit() {
    setBusy(true);
    const { ok, body } = await depositInvoice(invoiceId);
    setBusy(false);
    if (!ok || !body.success) {
      setError((body.message as string) ?? tLifecycle("deposit.error"));
      return;
    }
    setLifecycleRefresh((n) => n + 1);
    void load();
  }

  async function onConfirmDelete() {
    setBusy(true);
    const { ok, body } = await deleteDraftInvoice(invoiceId);
    setBusy(false);
    setDeleteDialogOpen(false);
    if (!ok || !body.success) {
      setError((body.message as string) ?? t("deleteError"));
      return;
    }
    router.push("/invoice");
  }

  if (loading) return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between gap-4">
        <Skeleton className="h-6 w-48" />
        <Skeleton className="h-9 w-32" />
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid grid-cols-2 gap-6">
          <Skeleton className="h-24 w-full" />
          <Skeleton className="h-24 w-full" />
        </div>
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-20 w-full" />
      </CardContent>
    </Card>
  );
  if (error && !invoice) return <p className="text-destructive">{error}</p>;
  if (!invoice) return <p className="text-destructive">{t("notFound")}</p>;

  const isIssued = invoice.status === "ISSUED" || invoice.status === "PAID";
  // Depositing on the platform is the NONE→DEPOSITED step (B6); the manual dialog
  // handles the later transitions, so the two controls never overlap.
  const canDeposit = isIssued && invoice.lifecycle_status === "NONE";
  const canAdvanceLifecycle =
    isIssued &&
    !canDeposit &&
    allowedNextLifecycleStatuses(invoice.lifecycle_status).length > 0;

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between gap-4">
        <CardTitle className="flex items-center gap-3">
          {invoice.invoice_number
            ? t("titleNumbered", { number: invoice.invoice_number })
            : t("titleDraft")}
          <InvoiceStatusBadge status={invoice.status} />
          <InvoiceLifecycleBadge status={invoice.lifecycle_status} />
        </CardTitle>
        <div className="flex items-center gap-2">
          <Button type="button" variant="outline" onClick={onDownload}>
            {t("downloadPdf")}
          </Button>
          {!isCustomer && invoice.status !== "DRAFT" ? (
            <Button type="button" variant="outline" onClick={onDownloadFacturx}>
              {t("downloadFacturx")}
            </Button>
          ) : null}
          {!isCustomer && invoice.status === "DRAFT" ? (
            <Button
              type="button"
              variant="destructive"
              onClick={() => setDeleteDialogOpen(true)}
              disabled={busy}
            >
              {t("deleteDraft")}
            </Button>
          ) : null}
          {!isCustomer && invoice.status === "ISSUED" ? (
            <Button type="button" variant="outline" onClick={onMarkPaid} disabled={busy}>
              {t("markPaid")}
            </Button>
          ) : null}
          {!isCustomer && isIssued ? (
            <Button
              type="button"
              variant="outline"
              onClick={() => setCreditDialogOpen(true)}
              disabled={busy}
            >
              {t("createCreditNote")}
            </Button>
          ) : null}
          {!isCustomer && canDeposit ? (
            <Button type="button" variant="outline" onClick={onDeposit} disabled={busy}>
              {tLifecycle("deposit.action")}
            </Button>
          ) : null}
          {!isCustomer && canAdvanceLifecycle ? (
            <Button
              type="button"
              variant="outline"
              onClick={() => setLifecycleDialogOpen(true)}
              disabled={busy}
            >
              {tLifecycle("advance")}
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

        {isIssued ? (
          <LinkedCreditNotes invoiceId={invoiceId} refreshKey={creditRefresh} />
        ) : null}

        {isIssued && !isCustomer ? (
          <LifecycleTimeline invoiceId={invoiceId} refreshKey={lifecycleRefresh} />
        ) : null}
      </CardContent>

      {!isCustomer && (
        <>
          <CreateCreditNoteDialog
            open={creditDialogOpen}
            onOpenChange={setCreditDialogOpen}
            invoice={invoice}
            onCreated={() => {
              setCreditRefresh((n) => n + 1);
              void load();
            }}
          />

          <AdvanceLifecycleDialog
            open={lifecycleDialogOpen}
            onOpenChange={setLifecycleDialogOpen}
            invoiceId={invoiceId}
            current={invoice.lifecycle_status}
            onApplied={() => {
              setLifecycleRefresh((n) => n + 1);
              void load();
            }}
          />

          <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>{t("deleteConfirmTitle")}</AlertDialogTitle>
                <AlertDialogDescription>
                  {t("deleteConfirmBody")}
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel disabled={busy}>
                  {t("deleteCancel")}
                </AlertDialogCancel>
                <AlertDialogAction
                  variant="destructive"
                  onClick={onConfirmDelete}
                  disabled={busy}
                >
                  {t("deleteConfirmAction")}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </>
      )}
    </Card>
  );
}
