"use client";

import { useTranslations } from "next-intl";
import { Badge } from "@/components/ui/badge";
import type { BackendInvoiceLifecycleStatus } from "@/types/backend";

const VARIANT: Record<
  Exclude<BackendInvoiceLifecycleStatus, "NONE">,
  "default" | "secondary" | "destructive" | "outline"
> = {
  DEPOSITED: "outline",
  RECEIVED: "secondary",
  APPROVED: "default",
  REJECTED: "destructive",
  COLLECTED: "secondary",
};

// Single authority on how each lifecycle status renders, including the "no
// platform lifecycle yet" case ("NONE"), shown as a muted dash (fr.json maps it).
export default function InvoiceLifecycleBadge({
  status,
}: {
  status: BackendInvoiceLifecycleStatus;
}) {
  const t = useTranslations("invoice.lifecycle.status");
  if (status === "NONE")
    return <span className="text-muted-foreground">{t("NONE")}</span>;
  return <Badge variant={VARIANT[status]}>{t(status)}</Badge>;
}
