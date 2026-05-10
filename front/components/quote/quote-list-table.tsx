"use client";

import { useEffect, useMemo, useState } from "react";
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
import { useMode } from "@/lib/mode-context";
import {
  type BackendQuote,
  type QuoteListStatus,
  quoteListStatus,
} from "@/types/backend";
import { PencilIcon } from "lucide-react";

type QuoteListItem = {
  id: string;
  projectName: string;
  status: QuoteListStatus;
};

export default function QuoteListTable() {
  const { isCustomer } = useMode();
  const [items, setItems] = useState<QuoteListItem[]>([]);

  const rowActions = useMemo<DataTableRowAction[]>(
    () => [
      {
        type: "link",
        label: isCustomer ? "Voir" : "Voir/Modifier",
        icon: PencilIcon,
        href: "/quote/{id}",
      },
    ],
    [isCustomer],
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
            status: quoteListStatus(quote),
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
          <DataTableSortableHead name="id">ID</DataTableSortableHead>
          <DataTableSortableHead name="projectName">
            Projet
          </DataTableSortableHead>
          <DataTableSortableHead name="status">Statut</DataTableSortableHead>
          <DataTableHead>Actions</DataTableHead>
        </DataTableRow>
      </DataTableHeader>
      <DataTableBody>
        {visibleItems.length === 0 ? (
          <DataTableRow>
            <DataTableCell className="text-muted-foreground">
              Aucun devis pour le moment.
            </DataTableCell>
            <DataTableCell> </DataTableCell>
            <DataTableCell> </DataTableCell>
            <DataTableCell> </DataTableCell>
          </DataTableRow>
        ) : (
          visibleItems.map((quote) => (
            <DataTableRow key={quote.id}>
              <DataTableCell>{quote.id}</DataTableCell>
              <DataTableCell>{quote.projectName}</DataTableCell>
              <DataTableCell>{quote.status}</DataTableCell>
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
