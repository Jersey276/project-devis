"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useTranslations } from "next-intl";
import { PlusIcon } from "lucide-react";
import { toast } from "sonner";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ClientsTable } from "./clients-table";
import { listClients } from "@/lib/services/clients";
import type { BackendClient } from "@/types/backend";

export default function ClientIndex() {
  const t = useTranslations("client.list");
  const [clients, setClients] = useState<BackendClient[]>([]);
  const cancelledRef = useRef(false);

  const reload = useCallback(async () => {
    const { ok, body } = await listClients();
    if (cancelledRef.current) return;
    if (ok && Array.isArray(body.clients)) {
      setClients(body.clients as BackendClient[]);
    } else if (!ok) {
      toast.error((body.message as string) ?? t("loadFailedToast"));
    }
  }, [t]);

  useEffect(() => {
    cancelledRef.current = false;
    reload();
    return () => {
      cancelledRef.current = true;
    };
  }, [reload]);

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between gap-4">
          <CardTitle>{t("title")}</CardTitle>
          <Button asChild>
            <Link
              href="/clients/create"
              className="inline-flex items-center gap-2"
            >
              <PlusIcon className="h-4 w-4" />
              {t("newButton")}
            </Link>
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <ClientsTable data={clients} onArchived={reload} />
      </CardContent>
    </Card>
  );
}
