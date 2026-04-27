"use client";

import { useEffect, useState } from "react";
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
import {
  type BackendQuote,
  type QuoteListStatus,
  quoteListStatus,
} from "@/types/backend";

type QuoteListItem = {
  id: string;
  projectName: string;
  status: QuoteListStatus;
};

const rowActions: DataTableRowAction[] = [
  {
    type: "link",
    label: "Voir/Modifier",
    href: "/quote/{id}",
  },
];

export default function QuoteListTable() {
  const [items, setItems] = useState<QuoteListItem[]>([]);

  useEffect(() => {
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
  }, []);

  return (
    <DataTable
      datas={items}
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
        {items.length === 0 ? (
          <DataTableRow>
            <DataTableCell className="text-muted-foreground">
              Aucun devis pour le moment.
            </DataTableCell>
            <DataTableCell> </DataTableCell>
            <DataTableCell> </DataTableCell>
            <DataTableCell> </DataTableCell>
          </DataTableRow>
        ) : (
          items.map((quote) => (
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
