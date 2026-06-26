"use client";

import { useTranslations } from "next-intl";
import { useAuth } from "@/lib/auth-context";
import { isSuperAdmin } from "@/lib/access";

type AdminGuardProps = {
  children: React.ReactNode;
};

export default function AdminGuard({ children }: AdminGuardProps) {
  const t = useTranslations("admin.guard");
  const { auth, ok } = useAuth();
  const allowed = ok && isSuperAdmin(auth);

  if (!allowed) {
    return (
      <div className="rounded-md border border-destructive/20 bg-destructive/5 p-4 text-sm text-destructive">
        {t("forbidden")}
      </div>
    );
  }

  return <>{children}</>;
}
