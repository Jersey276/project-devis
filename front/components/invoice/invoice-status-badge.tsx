"use client";

import { useTranslations } from "next-intl";
import { Badge } from "@/components/ui/badge";
import type { BackendInvoiceStatus } from "@/types/backend";

const VARIANT: Record<
  BackendInvoiceStatus,
  "default" | "secondary" | "destructive" | "outline"
> = {
  DRAFT: "outline",
  ISSUED: "default",
  PAID: "secondary",
  CANCELLED: "destructive",
};

export default function InvoiceStatusBadge({
  status,
}: {
  status: BackendInvoiceStatus;
}) {
  const t = useTranslations("invoice.status");
  return <Badge variant={VARIANT[status]}>{t(status)}</Badge>;
}
