"use client";

import { useEffect, useMemo, useState } from "react";
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
import { listQuotes } from "@/lib/services/quotes";
import { exportQuotePdf } from "@/lib/services/export";
import { useMode } from "@/lib/mode-context";
import { formatEurosFromCents } from "@/lib/utils";
import {
  type BackendQuote,
  type QuoteListState,
  quoteListState,
} from "@/types/backend";
import { DownloadIcon, PencilIcon } from "lucide-react";
import { toast } from "sonner";

type QuoteListItem = {
  id: string;
  projectName: string;
  status: QuoteListState;
  totalTtc: number;
};

export default function QuoteListTable() {
  const { isCustomer } = useMode();
  const t = useTranslations("quote.list");
  const tStatus = useTranslations("status.quote");
  const [items, setItems] = useState<QuoteListItem[]>([]);

  const rowActions = useMemo<DataTableRowAction[]>(
    () => [
      {
        type: "link",
        label: isCustomer ? t("actions.view") : t("actions.viewEdit"),
        icon: PencilIcon,
        href: "/quote/{id}",
      },
      {
        type: "callback",
        label: t("actions.exportPdf"),
        icon: DownloadIcon,
        hidden: isCustomer,
        callback: (row) => {
          const id = (row as { id: string }).id;
          exportQuotePdf(id).catch((err) => {
            console.error("export quote pdf failed", err);
            toast.error(t("exportFailedToast"));
          });
        },
      },
    ],
    [isCustomer, t],
  );

  useEffect(() => {
    // Customer mode has no client→quote relation yet (step 2). Skip the fetch
    // entirely so the empty state renders without leaking provider data.
    if (isCustomer) return;
    let cancelled = false;
    listQuotes().then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && Array.isArray(body.quotes)) {
        const quotes = body.quotes as BackendQuote[];
        setItems(
          quotes.map((quote) => ({
            id: quote.quote_id,
            projectName: quote.name,
            status: quoteListState(quote),
            totalTtc: quote.total_ttc ?? 0,
          })),
        );
      }
    });
    return () => {
      cancelled = true;
    };
  }, [isCustomer]);

  // Hide any quotes fetched in provider mode if the user toggles to customer.
  const visibleItems = isCustomer ? [] : items;

  return (
    <DataTable
      datas={visibleItems}
      sortBy="id"
      sortDirection="asc"
      row_actions={rowActions}
    >
      <DataTableHeader>
        <DataTableRow>
          <DataTableSortableHead name="id">{t("columns.id")}</DataTableSortableHead>
          <DataTableSortableHead name="projectName">
            {t("columns.project")}
          </DataTableSortableHead>
          <DataTableSortableHead name="status">{t("columns.status")}</DataTableSortableHead>
          <DataTableSortableHead name="totalTtc">
            {t("columns.totalTtc")}
          </DataTableSortableHead>
          <DataTableHead>{t("columns.actions")}</DataTableHead>
        </DataTableRow>
      </DataTableHeader>
      <DataTableBody>
        {visibleItems.length === 0 ? (
          <DataTableRow>
            <DataTableCell className="text-muted-foreground">
              {t("empty")}
            </DataTableCell>
            <DataTableCell> </DataTableCell>
            <DataTableCell> </DataTableCell>
            <DataTableCell> </DataTableCell>
            <DataTableCell> </DataTableCell>
          </DataTableRow>
        ) : (
          visibleItems.map((quote) => (
            <DataTableRow key={quote.id}>
              <DataTableCell>{quote.id}</DataTableCell>
              <DataTableCell>{quote.projectName}</DataTableCell>
              <DataTableCell>{tStatus(quote.status)}</DataTableCell>
              <DataTableCell className="tabular-nums">
                {formatEurosFromCents(quote.totalTtc)}
              </DataTableCell>
              <DataTableCell>
                <DataTableRowActions id={quote.id} row={quote} />
              </DataTableCell>
            </DataTableRow>
          ))
        )}
      </DataTableBody>
    </DataTable>
  );
}
