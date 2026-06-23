"use client";

import { Suspense, useCallback, useEffect, useState } from "react";
import { useSearchParams, useRouter, usePathname } from "next/navigation";
import Link from "next/link";
import { useTranslations } from "next-intl";
import { PlusIcon } from "lucide-react";
import { toast } from "sonner";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { FilterSidebar, FilterSidebarSection } from "@/components/ui/filter-sidebar";
import { SelectCombobox } from "@/components/ui/select-combobox";
import { ClientsTable } from "./clients-table";
import { listClients } from "@/lib/services/clients";
import type { BackendClient } from "@/types/backend";

const PAGE_SIZE = 20;

function ClientIndex() {
  const t = useTranslations("client.list");
  const tCommon = useTranslations("common.filterSidebar");
  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();

  const page = Number(searchParams.get("page") ?? "1");
  const search = searchParams.get("search") ?? "";
  const clientTypes = searchParams.get("client_types")
    ? searchParams.get("client_types")!.split(",")
    : [];

  const [clients, setClients] = useState<BackendClient[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);

  function pushParams(newSearch: string, newTypes: string[], newPage: number) {
    const p = new URLSearchParams();
    if (newPage > 1) p.set("page", String(newPage));
    if (newSearch) p.set("search", newSearch);
    if (newTypes.length > 0) p.set("client_types", newTypes.join(","));
    router.push(`${pathname}?${p.toString()}`);
  }

  const fetchClients = useCallback(async () => {
    setLoading(true);
    const params = new URLSearchParams({
      page: String(page),
      page_size: String(PAGE_SIZE),
    });
    if (search) params.set("search", search);
    if (clientTypes.length > 0) params.set("client_types", clientTypes.join(","));

    const { ok, body } = await listClients(params.toString());
    if (ok && Array.isArray(body.clients)) {
      setClients(body.clients as BackendClient[]);
      setTotal((body.total ?? 0) as number);
    } else if (!ok) {
      toast.error((body.message as string) ?? t("loadFailedToast"));
    }
    setLoading(false);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams]);

  useEffect(() => { void fetchClients(); }, [fetchClients]);

  const reload = useCallback(() => { void fetchClients(); }, [fetchClients]);

  const activeFilterCount = (search ? 1 : 0) + (clientTypes.length > 0 ? 1 : 0);
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  const tForm = useTranslations("client.form.clientType");
  const CLIENT_TYPE_ITEMS = [
    { value: "individual", label: tForm("individual") },
    { value: "business", label: tForm("business") },
  ];

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between gap-4">
          <CardTitle>{t("title")}</CardTitle>
          <Button asChild>
            <Link href="/clients/create" className="inline-flex items-center gap-2">
              <PlusIcon className="h-4 w-4" />
              {t("newButton")}
            </Link>
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex flex-wrap items-center gap-2">
          <Input
            className="w-full sm:w-64"
            placeholder={t("filters.searchPlaceholder")}
            value={search}
            onChange={(e) => pushParams(e.target.value, clientTypes, 1)}
          />
          <FilterSidebar
            triggerLabel={tCommon("trigger")}
            title={tCommon("title")}
            resetLabel={tCommon("reset")}
            activeCount={activeFilterCount - (search ? 1 : 0)}
            onReset={() => pushParams(search, [], 1)}
          >
            <FilterSidebarSection label={t("filters.typeLabel")}>
              <SelectCombobox
                multiple
                items={CLIENT_TYPE_ITEMS}
                value={clientTypes}
                onValueChange={(vals) => pushParams(search, vals, 1)}
                placeholder={t("filters.typePlaceholder")}
                emptyLabel={t("filters.typeEmpty")}
              />
            </FilterSidebarSection>
          </FilterSidebar>
        </div>

        {loading ? (
          <p className="text-sm text-muted-foreground">{t("loadFailedToast")}</p>
        ) : (
          <>
            <ClientsTable data={clients} onArchived={reload} />
            {total > 0 && (
              <div className="flex flex-wrap items-center justify-between gap-2 text-sm text-muted-foreground">
                <span>{total} client{total > 1 ? "s" : ""}</span>
                <div className="flex gap-2">
                  <button
                    className="rounded border px-3 py-1 disabled:opacity-40"
                    disabled={page <= 1}
                    onClick={() => pushParams(search, clientTypes, page - 1)}
                  >
                    ←
                  </button>
                  <span>{page} / {totalPages}</span>
                  <button
                    className="rounded border px-3 py-1 disabled:opacity-40"
                    disabled={page >= totalPages}
                    onClick={() => pushParams(search, clientTypes, page + 1)}
                  >
                    →
                  </button>
                </div>
              </div>
            )}
          </>
        )}
      </CardContent>
    </Card>
  );
}

export default function ClientIndexPage() {
  return (
    <Suspense fallback={null}>
      <ClientIndex />
    </Suspense>
  );
}
