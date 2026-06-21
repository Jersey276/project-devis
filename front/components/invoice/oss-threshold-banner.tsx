"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { getOSSThresholdStatus } from "@/lib/services/invoices";
import { formatEurosFromCents } from "@/lib/utils";
import type { BackendOSSThresholdStatus } from "@/types/backend";

// Shows where the issuer stands against the OSS distance-selling threshold
// (art. 259 D CGI) for the current civil year. Hidden for purely domestic
// sellers (no OSS turnover and not opted in) to avoid cluttering the page.
export default function OSSThresholdBanner() {
  const t = useTranslations("invoice.oss");
  const [status, setStatus] = useState<BackendOSSThresholdStatus | null>(null);

  useEffect(() => {
    let cancelled = false;
    getOSSThresholdStatus().then(({ ok, body }) => {
      if (cancelled || !ok || !body.success) return;
      setStatus(body as unknown as BackendOSSThresholdStatus);
    });
    return () => {
      cancelled = true;
    };
  }, []);

  if (!status) return null;
  if (!status.oss_active && status.cumulative_ht_cents <= 0) return null;

  const remaining = Math.max(0, status.threshold_cents - status.cumulative_ht_cents);
  const pct = Math.min(
    100,
    status.threshold_cents > 0
      ? Math.round((status.cumulative_ht_cents / status.threshold_cents) * 100)
      : 0,
  );

  return (
    <Alert className="mb-4" variant={status.oss_active ? "destructive" : "default"}>
      <AlertTitle>{t("title", { year: status.year })}</AlertTitle>
      <AlertDescription>
        <p>
          {t("cumulative", {
            cumulative: formatEurosFromCents(status.cumulative_ht_cents),
            threshold: formatEurosFromCents(status.threshold_cents),
            pct,
          })}
        </p>
        <p>
          {status.oss_active
            ? status.oss_enabled
              ? t("activeOptIn")
              : t("activeThreshold")
            : t("remaining", { remaining: formatEurosFromCents(remaining) })}
        </p>
      </AlertDescription>
    </Alert>
  );
}
