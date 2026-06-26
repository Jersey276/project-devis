"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { canUsePaidFeatures } from "@/lib/access";
import { useMode } from "@/lib/mode-context";
import { useAuth } from "@/lib/auth-context";
import { apiFetch } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";
import type { AuthContext } from "@/lib/access";

type SubscriptionGuardProps = {
  children: React.ReactNode;
};

export default function SubscriptionGuard({
  children,
}: SubscriptionGuardProps) {
  const t = useTranslations("subscription.guard");
  const { isCustomer } = useMode();
  const { auth: ssrAuth, ok: ssrOk } = useAuth();

  const [clientAllowed, setClientAllowed] = useState<boolean | null>(
    ssrOk ? canUsePaidFeatures(ssrAuth) : null,
  );

  useEffect(() => {
    if (ssrOk) return;
    let cancelled = false;
    apiFetch("/api/auth/me").then(({ ok, body }) => {
      if (cancelled) return;
      const auth = (body?.auth ?? null) as AuthContext | null;
      setClientAllowed(ok && body?.success === true && canUsePaidFeatures(auth));
    });
    return () => {
      cancelled = true;
    };
  }, [ssrOk]);

  if (isCustomer || clientAllowed === true) return <>{children}</>;

  if (clientAllowed === null) {
    return <Skeleton className="h-4 w-32" />;
  }

  return (
    <div className="rounded-md border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800 dark:border-amber-800 dark:bg-amber-950 dark:text-amber-200">
      {t("forbidden")}
    </div>
  );
}
