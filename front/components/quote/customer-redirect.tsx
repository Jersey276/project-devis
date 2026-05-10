"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useMode } from "@/lib/mode-context";

export default function CustomerRedirect({ to }: { to: string }) {
  const router = useRouter();
  const { isCustomer } = useMode();
  useEffect(() => {
    if (isCustomer) router.replace(to);
  }, [isCustomer, router, to]);
  return null;
}
