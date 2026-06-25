"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { apiFetch } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";
import type { AuthContext } from "@/lib/access";
import { isSuperAdmin } from "@/lib/access";

type AdminGuardProps = {
  children: React.ReactNode;
};

export default function AdminGuard({ children }: AdminGuardProps) {
  const t = useTranslations("admin.guard");
  const [loading, setLoading] = useState(true);
  const [allowed, setAllowed] = useState(false);

  useEffect(() => {
    let cancelled = false;

    apiFetch("/api/auth/me").then(({ ok, body }) => {
      if (cancelled) return;
      const auth = (body.auth ?? null) as AuthContext | null;
      setAllowed(ok && body.success === true && isSuperAdmin(auth));
      setLoading(false);
    });

    return () => {
      cancelled = true;
    };
  }, []);

  if (loading) {
    return <Skeleton className="h-4 w-32" />;
  }

  if (!allowed) {
    return (
      <div className="rounded-md border border-destructive/20 bg-destructive/5 p-4 text-sm text-destructive">
        {t("forbidden")}
      </div>
    );
  }

  return <>{children}</>;
}
