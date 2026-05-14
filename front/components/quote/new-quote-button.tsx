"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { useMode } from "@/lib/mode-context";

export default function NewQuoteButton() {
  const { isCustomer } = useMode();
  const t = useTranslations("quote.list");
  if (isCustomer) return null;
  return (
    <Button asChild>
      <Link href="/quote/create">{t("newButton")}</Link>
    </Button>
  );
}
