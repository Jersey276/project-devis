"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  getCreditNote,
  readCreditNoteFromBody,
} from "@/lib/services/invoices";
import { exportCreditNotePdf } from "@/lib/services/export";
import { formatEurosFromCents } from "@/lib/utils";
import type {
  BackendCreditNoteDetails,
  BackendInvoiceParty,
} from "@/types/backend";

function partyLines(p: BackendInvoiceParty | undefined): string[] {
  if (!p) return [];
  const lines: string[] = [];
  const title = p.company || `${p.first_name} ${p.last_name}`.trim();
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

export default function CreditNoteDetail({
  creditNoteId,
}: {
  creditNoteId: string;
}) {
  const t = useTranslations("creditNote.detail");
  const [cn, setCn] = useState<BackendCreditNoteDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    getCreditNote(creditNoteId).then(({ ok, body }) => {
      if (cancelled) return;
      setLoading(false);
      if (!ok || !body.success) {
        setError((body.message as string) ?? t("loadError"));
        return;
      }
      setError(null);
      setCn(readCreditNoteFromBody(body));
    });
    return () => {
      cancelled = true;
    };
  }, [creditNoteId, t]);

  async function onDownload() {
    try {
      await exportCreditNotePdf(creditNoteId);
    } catch {
      setError(t("pdfError"));
    }
  }

  if (loading) return <p>{t("loading")}</p>;
  if (error && !cn) return <p className="text-destructive">{error}</p>;
  if (!cn) return <p className="text-destructive">{t("notFound")}</p>;

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between gap-4">
        <CardTitle>{t("title", { number: cn.credit_note_number })}</CardTitle>
        <Button type="button" variant="outline" onClick={onDownload}>
          {t("downloadPdf")}
        </Button>
      </CardHeader>
      <CardContent className="space-y-6">
        {error ? <p className="text-sm text-destructive">{error}</p> : null}

        <div className="text-sm text-muted-foreground">
          {t("originInvoice", { number: cn.invoice_number })}
          {cn.reason ? ` — ${cn.reason}` : ""}
        </div>

        <div className="grid grid-cols-2 gap-6 text-sm">
          <section>
            <h3 className="mb-1 font-semibold">{t("issuer")}</h3>
            {partyLines(cn.issuer).map((line, i) => (
              <div key={i}>{line}</div>
            ))}
          </section>
          <section>
            <h3 className="mb-1 font-semibold">{t("client")}</h3>
            {partyLines(cn.client).map((line, i) => (
              <div key={i}>{line}</div>
            ))}
          </section>
        </div>

        <table className="w-full border-collapse text-sm">
          <thead>
            <tr className="border-b text-left">
              <th className="py-1">{t("lines.name")}</th>
              <th className="py-1 text-center">{t("lines.quantity")}</th>
              <th className="py-1 text-center">{t("lines.unit")}</th>
              <th className="py-1 text-right">{t("lines.unitPrice")}</th>
              {!cn.vat_exempt ? (
                <th className="py-1 text-center">{t("lines.vat")}</th>
              ) : null}
              <th className="py-1 text-right">{t("lines.totalHt")}</th>
            </tr>
          </thead>
          <tbody>
            {cn.lines.map((l, i) => (
              <tr key={i} className="border-b">
                <td className="py-1">{l.name}</td>
                <td className="py-1 text-center">{l.quantity}</td>
                <td className="py-1 text-center">{l.unit}</td>
                <td className="py-1 text-right tabular-nums">
                  {formatEurosFromCents(l.unit_price_cents)}
                </td>
                {!cn.vat_exempt ? (
                  <td className="py-1 text-center">{l.tax_rate} %</td>
                ) : null}
                <td className="py-1 text-right tabular-nums">
                  -{formatEurosFromCents(l.line_ht_cents)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        <div className="ml-auto w-64 space-y-1 text-sm">
          <div className="flex justify-between">
            <span>{t("totals.ht")}</span>
            <span className="tabular-nums">
              -{formatEurosFromCents(cn.total_ht_cents)}
            </span>
          </div>
          {!cn.vat_exempt ? (
            <div className="flex justify-between">
              <span>{t("totals.vat")}</span>
              <span className="tabular-nums">
                -{formatEurosFromCents(cn.total_vat_cents)}
              </span>
            </div>
          ) : null}
          <div className="flex justify-between border-t pt-1 font-semibold">
            <span>{t("totals.ttc")}</span>
            <span className="tabular-nums">
              -{formatEurosFromCents(cn.total_ttc_cents)}
            </span>
          </div>
        </div>

        {cn.vat_exempt ? (
          <p className="text-xs text-muted-foreground">{t("vatExemptNotice")}</p>
        ) : null}
      </CardContent>
    </Card>
  );
}
