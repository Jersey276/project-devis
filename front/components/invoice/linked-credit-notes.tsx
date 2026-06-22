"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import {
  listCreditNotes,
  readCreditNotesFromBody,
} from "@/lib/services/invoices";
import { exportCreditNotePdf } from "@/lib/services/export";
import { formatEurosFromCents } from "@/lib/utils";
import type { BackendCreditNoteSummary } from "@/types/backend";

type Props = {
  invoiceId: string;
  refreshKey?: number;
};

export default function LinkedCreditNotes({ invoiceId, refreshKey }: Props) {
  const t = useTranslations("creditNote.linkedSection");
  const [items, setItems] = useState<BackendCreditNoteSummary[]>([]);

  useEffect(() => {
    let cancelled = false;
    listCreditNotes(invoiceId).then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && body.success) setItems(readCreditNotesFromBody(body));
    });
    return () => {
      cancelled = true;
    };
  }, [invoiceId, refreshKey]);

  if (items.length === 0) return null;

  return (
    <section className="space-y-2">
      <h3 className="font-semibold">{t("title")}</h3>
      <table className="w-full border-collapse text-sm">
        <tbody>
          {items.map((cn) => (
            <tr key={cn.credit_note_id} className="border-b">
              <td className="py-1">
                <Link
                  href={`/credit-note/${cn.credit_note_id}`}
                  className="underline"
                >
                  {cn.credit_note_number}
                </Link>
              </td>
              <td className="py-1 text-muted-foreground">
                {cn.is_total ? t("total") : t("partial")}
              </td>
              <td className="py-1 text-right tabular-nums">
                -{formatEurosFromCents(cn.total_ttc_cents)}
              </td>
              <td className="py-1 text-right">
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    void exportCreditNotePdf(cn.credit_note_id);
                  }}
                >
                  {t("downloadPdf")}
                </Button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}
