"use client";

import { Suspense, useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams, useRouter, usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import {
  DataTable,
  DataTableBody,
  DataTableCell,
  DataTableHead,
  DataTableHeader,
  DataTableRow,
  DataTableRowAction,
  DataTableRowActions,
  DataTableSortableHead,
} from "@/components/custom/data-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { FilterSidebar, FilterSidebarSection } from "@/components/ui/filter-sidebar";
import { SelectCombobox } from "@/components/ui/select-combobox";
import { DateRangePicker } from "@/components/ui/date-range-picker";
import {
  listCreditNotes,
  readCreditNotesFromBody,
} from "@/lib/services/invoices";
import { exportCreditNotePdf } from "@/lib/services/export";
import { formatEurosFromCents } from "@/lib/utils";
import type { BackendCreditNoteSummary } from "@/types/backend";

const PAGE_SIZE = 20;

type CreditNoteRow = {
  id: string;
  number: string;
  invoiceNumber: string;
  issuedAt: string;
  isTotal: boolean;
  totalTtc: number;
};

function toRows(items: BackendCreditNoteSummary[]): CreditNoteRow[] {
  return items.map((cn) => ({
    id: cn.credit_note_id,
    number: cn.credit_note_number,
    invoiceNumber: cn.invoice_number,
    issuedAt: cn.issued_at,
    isTotal: cn.is_total,
    totalTtc: cn.total_ttc_cents,
  }));
}

function CreditNoteListTableInner() {
  const t = useTranslations("creditNote.list");
  const tFilters = useTranslations("creditNote.list.filters");
  const tCommon = useTranslations("common.filterSidebar");

  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();

  const page = Number(searchParams.get("page") ?? "1");
  const isTotal = searchParams.get("is_total") ?? "";
  const issuedFrom = searchParams.get("issued_from") ?? "";
  const issuedTo = searchParams.get("issued_to") ?? "";

  const [items, setItems] = useState<CreditNoteRow[]>([]);
  const [total, setTotal] = useState(0);
  const [error, setError] = useState<string | null>(null);

  function pushParams(p: {
    isTotal?: string;
    issuedFrom?: string;
    issuedTo?: string;
    page?: number;
  }) {
    const next = new URLSearchParams();
    const pg = p.page ?? 1;
    const it = p.isTotal ?? isTotal;
    const iF = p.issuedFrom ?? issuedFrom;
    const iT = p.issuedTo ?? issuedTo;
    if (pg > 1) next.set("page", String(pg));
    if (it) next.set("is_total", it);
    if (iF) next.set("issued_from", iF);
    if (iT) next.set("issued_to", iT);
    router.push(`${pathname}?${next.toString()}`);
  }

  const fetchCreditNotes = useCallback(async () => {
    const params = new URLSearchParams({ page: String(page), page_size: String(PAGE_SIZE) });
    if (isTotal) params.set("is_total", isTotal);
    if (issuedFrom) params.set("issued_from", issuedFrom);
    if (issuedTo) params.set("issued_to", issuedTo);

    const { ok, body } = await listCreditNotes(params.toString());
    if (!ok || !body.success) {
      setError((body.message as string) ?? t("loadError"));
      return;
    }
    setError(null);
    setItems(toRows(readCreditNotesFromBody(body)));
    setTotal((body.total ?? 0) as number);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams]);

  useEffect(() => {
    void fetchCreditNotes();
  }, [fetchCreditNotes]);

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  const isTotalItems = [
    { value: "true", label: t("total") },
    { value: "false", label: t("partial") },
  ];

  const rowActions = useMemo<DataTableRowAction[]>(
    () => [
      { type: "link", label: t("actions.open"), href: "/credit-note/{id}" },
      {
        type: "callback",
        label: t("actions.downloadPdf"),
        callback: (row) => {
          void exportCreditNotePdf((row as CreditNoteRow).id).catch(() =>
            setError(t("pdfError")),
          );
        },
      },
    ],
    [t],
  );

  const hasFilters = !!isTotal || !!issuedFrom || !!issuedTo;

  return (
    <>
      {error ? <p className="mb-4 text-sm text-destructive">{error}</p> : null}

      <div className="flex items-start gap-4">
        <FilterSidebar
          triggerLabel={tCommon("trigger")}
          title={tCommon("title")}
          activeCount={hasFilters ? 1 : 0}
          onReset={() => pushParams({ isTotal: "", issuedFrom: "", issuedTo: "", page: 1 })}
          resetLabel={tCommon("reset")}
        >
          <FilterSidebarSection label={tFilters("typeLabel")}>
            <SelectCombobox
              items={isTotalItems}
              value={isTotal}
              onValueChange={(v) => pushParams({ isTotal: v, page: 1 })}
              placeholder={tFilters("typePlaceholder")}
              emptyLabel={tFilters("typeEmpty")}
            />
          </FilterSidebarSection>

          <FilterSidebarSection label={tFilters("issuedDateLabel")}>
            <DateRangePicker
              from={issuedFrom}
              to={issuedTo}
              onValueChange={(from, to) => pushParams({ issuedFrom: from, issuedTo: to, page: 1 })}
            />
          </FilterSidebarSection>
        </FilterSidebar>

        <div className="flex-1 min-w-0">
          <DataTable
            datas={items}
            sortBy="number"
            sortDirection="desc"
            row_actions={rowActions}
          >
            <DataTableHeader>
              <DataTableRow>
                <DataTableSortableHead name="number">
                  {t("columns.number")}
                </DataTableSortableHead>
                <DataTableSortableHead name="invoiceNumber">
                  {t("columns.invoice")}
                </DataTableSortableHead>
                <DataTableSortableHead name="isTotal">
                  {t("columns.type")}
                </DataTableSortableHead>
                <DataTableSortableHead name="issuedAt">
                  {t("columns.issuedAt")}
                </DataTableSortableHead>
                <DataTableSortableHead name="totalTtc">
                  {t("columns.totalTtc")}
                </DataTableSortableHead>
                <DataTableHead>{t("columns.actions")}</DataTableHead>
              </DataTableRow>
            </DataTableHeader>
            <DataTableBody>
              {items.length === 0 ? (
                <DataTableRow>
                  <DataTableCell className="text-muted-foreground">
                    {t("empty")}
                  </DataTableCell>
                  <DataTableCell> </DataTableCell>
                  <DataTableCell> </DataTableCell>
                  <DataTableCell> </DataTableCell>
                  <DataTableCell> </DataTableCell>
                  <DataTableCell> </DataTableCell>
                </DataTableRow>
              ) : (
                items.map((item) => (
                  <DataTableRow key={item.id}>
                    <DataTableCell>{item.number}</DataTableCell>
                    <DataTableCell>{item.invoiceNumber || "—"}</DataTableCell>
                    <DataTableCell>
                      <Badge variant={item.isTotal ? "default" : "secondary"}>
                        {item.isTotal ? t("total") : t("partial")}
                      </Badge>
                    </DataTableCell>
                    <DataTableCell>{item.issuedAt || "—"}</DataTableCell>
                    <DataTableCell className="tabular-nums">
                      -{formatEurosFromCents(item.totalTtc)}
                    </DataTableCell>
                    <DataTableCell>
                      <DataTableRowActions id={item.id} row={item} />
                    </DataTableCell>
                  </DataTableRow>
                ))
              )}
            </DataTableBody>
          </DataTable>

          {totalPages > 1 && (
            <div className="mt-4 flex items-center justify-end gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={page <= 1}
                onClick={() => pushParams({ page: page - 1 })}
              >
                Précédent
              </Button>
              <span className="text-sm text-muted-foreground">
                {page} / {totalPages}
              </span>
              <Button
                variant="outline"
                size="sm"
                disabled={page >= totalPages}
                onClick={() => pushParams({ page: page + 1 })}
              >
                Suivant
              </Button>
            </div>
          )}
        </div>
      </div>
    </>
  );
}

export default function CreditNoteListTable() {
  return (
    <Suspense fallback={null}>
      <CreditNoteListTableInner />
    </Suspense>
  );
}
