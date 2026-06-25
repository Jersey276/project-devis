"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { apiFetch } from "@/lib/api";
import { canUsePaidFeatures, type AuthContext } from "@/lib/access";
import { useMode } from "@/lib/mode-context";

type SubscriptionGuardProps = {
  children: React.ReactNode;
};

export default function SubscriptionGuard({
  children,
}: SubscriptionGuardProps) {
  const t = useTranslations("subscription.guard");
  const { isCustomer } = useMode();
  const [loading, setLoading] = useState(true);
  const [allowed, setAllowed] = useState(false);

  useEffect(() => {
    if (isCustomer) return;

    let cancelled = false;

    apiFetch("/api/auth/me").then(({ ok, body }) => {
      if (cancelled) return;
      const auth = (body.auth ?? null) as AuthContext | null;
      setAllowed(ok && body.success === true && canUsePaidFeatures(auth));
      setLoading(false);
    });

    return () => {
      cancelled = true;
    };
  }, [isCustomer]);

  if (isCustomer) return <>{children}</>;

  if (loading) {
    return <p className="text-sm text-muted-foreground">{t("loading")}</p>;
  }

  if (!allowed) {
    return (
      <div className="rounded-md border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800 dark:border-amber-800 dark:bg-amber-950 dark:text-amber-200">
        {t("forbidden")}
      </div>
    );
  }

  return <>{children}</>;
}
