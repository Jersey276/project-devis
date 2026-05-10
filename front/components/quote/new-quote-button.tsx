"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { useMode } from "@/lib/mode-context";

export default function NewQuoteButton() {
  const { isCustomer } = useMode();
  if (isCustomer) return null;
  return (
    <Button asChild>
      <Link href="/quote/create">Nouveau devis</Link>
    </Button>
  );
}
