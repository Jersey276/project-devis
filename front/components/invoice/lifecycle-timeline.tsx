"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  listInvoiceLifecycleEvents,
  readLifecycleEventsFromBody,
} from "@/lib/services/invoices";
import type { BackendInvoiceLifecycleEvent } from "@/types/backend";

type Props = {
  invoiceId: string;
  refreshKey?: number;
};

export default function LifecycleTimeline({ invoiceId, refreshKey }: Props) {
  const t = useTranslations("invoice.lifecycle");
  const [events, setEvents] = useState<BackendInvoiceLifecycleEvent[]>([]);

  useEffect(() => {
    let cancelled = false;
    listInvoiceLifecycleEvents(invoiceId).then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && body.success) setEvents(readLifecycleEventsFromBody(body));
    });
    return () => {
      cancelled = true;
    };
  }, [invoiceId, refreshKey]);

  return (
    <section className="space-y-2">
      <h3 className="font-semibold">{t("history.title")}</h3>
      {events.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("history.empty")}</p>
      ) : (
        <ol className="space-y-1 text-sm">
          {events.map((e, i) => (
            <li key={i} className="flex items-baseline gap-2 border-b py-1">
              <span className="font-medium">{t(`status.${e.status}`)}</span>
              <span className="text-muted-foreground">{e.created_at}</span>
              {e.note ? <span className="text-muted-foreground">— {e.note}</span> : null}
            </li>
          ))}
        </ol>
      )}
    </section>
  );
}
