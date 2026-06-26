"use client";

import { useTranslations } from "next-intl";
import { canUsePaidFeatures } from "@/lib/access";
import { useMode } from "@/lib/mode-context";
import { useAuth } from "@/lib/auth-context";

type SubscriptionGuardProps = {
  children: React.ReactNode;
};

export default function SubscriptionGuard({
  children,
}: SubscriptionGuardProps) {
  const t = useTranslations("subscription.guard");
  const { isCustomer } = useMode();
  const { auth, ok } = useAuth();

  if (isCustomer) return <>{children}</>;

  const allowed = ok && canUsePaidFeatures(auth);

  if (!allowed) {
    return (
      <div className="rounded-md border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800 dark:border-amber-800 dark:bg-amber-950 dark:text-amber-200">
        {t("forbidden")}
      </div>
    );
  }

  return <>{children}</>;
}
