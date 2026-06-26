"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { apiFetch } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";
import { useAuth } from "@/lib/auth-context";
import { isSuperAdmin } from "@/lib/access";
import type { AuthContext } from "@/lib/access";

type AdminGuardProps = {
  children: React.ReactNode;
};

export default function AdminGuard({ children }: AdminGuardProps) {
  const t = useTranslations("admin.guard");
  const { auth: ssrAuth, ok: ssrOk } = useAuth();

  // When SSR auth is available (gateway reachable), use it synchronously.
  // When SSR auth is missing (gateway unreachable during Next.js render, e.g.
  // in Cypress test env), fall back to a client-side fetch so cy.intercept works.
  const [clientAllowed, setClientAllowed] = useState<boolean | null>(
    ssrOk ? isSuperAdmin(ssrAuth) : null,
  );

  useEffect(() => {
    if (ssrOk) return; // SSR result is authoritative — no client fetch needed.
    let cancelled = false;
    apiFetch("/api/auth/me").then(({ ok, body }) => {
      if (cancelled) return;
      const auth = (body?.auth ?? null) as AuthContext | null;
      setClientAllowed(ok && body?.success === true && isSuperAdmin(auth));
    });
    return () => {
      cancelled = true;
    };
  }, [ssrOk]);

  if (clientAllowed === null) {
    return <Skeleton className="h-4 w-32" />;
  }

  if (!clientAllowed) {
    return (
      <div className="rounded-md border border-destructive/20 bg-destructive/5 p-4 text-sm text-destructive">
        {t("forbidden")}
      </div>
    );
  }

  return <>{children}</>;
}
