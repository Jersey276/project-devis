"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { createInvoiceFromQuote } from "@/lib/services/invoices";
import { listSchedules } from "@/lib/services/schedules";
import { readSchedulesFromBody } from "@/lib/services/schedules";

type Props = {
  quoteId: string;
  /** Whether the parent quote is validated; the button only shows when true. */
  validated: boolean;
  onError?: (message: string) => void;
};

// Whole-quote invoicing is only offered when the quote is validated AND has no
// schedule. If a schedule exists the user must bill from it, so we hide the
// button (the backend also rejects this case with QuoteHasSchedule).
export default function GenerateInvoiceFromQuoteButton({
  quoteId,
  validated,
  onError,
}: Props) {
  const t = useTranslations("invoice.generateFromQuote");
  const router = useRouter();
  const [hasSchedule, setHasSchedule] = useState<boolean | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!validated) return;
    let cancelled = false;
    listSchedules(`quote_id=${encodeURIComponent(quoteId)}`).then(({ ok, body }) => {
      if (cancelled) return;
      if (!ok || !body.success) {
        setHasSchedule(false);
        return;
      }
      setHasSchedule(readSchedulesFromBody(body).length > 0);
    });
    return () => {
      cancelled = true;
    };
  }, [quoteId, validated]);

  if (!validated || hasSchedule !== false) return null;

  async function onClick() {
    setSubmitting(true);
    const { ok, body } = await createInvoiceFromQuote({
      quoteId,
      issueNow: true,
    });
    setSubmitting(false);
    if (!ok || !body.success) {
      onError?.((body.message as string) ?? t("error"));
      return;
    }
    const invoiceId = (body.invoice_id as string) ?? "";
    if (invoiceId) router.push(`/invoice/${invoiceId}`);
  }

  return (
    <Button type="button" variant="outline" onClick={onClick} disabled={submitting}>
      {submitting ? t("submitting") : t("generate")}
    </Button>
  );
}
