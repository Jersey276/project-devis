"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
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
import InvoiceStatusBadge from "@/components/invoice/invoice-status-badge";
import {
  cancelInvoice,
  listInvoices,
  markInvoicePaid,
  readInvoicesFromBody,
} from "@/lib/services/invoices";
import { exportInvoicePdf } from "@/lib/services/export";
import { formatEurosFromCents } from "@/lib/utils";
import type {
  BackendInvoiceStatus,
  BackendInvoiceSummary,
} from "@/types/backend";

type InvoiceRow = {
  id: string;
  number: string;
  status: BackendInvoiceStatus;
  quoteId: string;
  dueDate: string;
  totalTtc: number;
};

function toRows(invoices: BackendInvoiceSummary[]): InvoiceRow[] {
  return invoices.map((i) => ({
    id: i.invoice_id,
    number: i.invoice_number,
    status: i.status,
    quoteId: i.quote_id,
    dueDate: i.due_date,
    totalTtc: i.total_ttc_cents,
  }));
}

export default function InvoiceListTable() {
  const t = useTranslations("invoice.list");
  const [items, setItems] = useState<InvoiceRow[]>([]);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    const { ok, body } = await listInvoices();
    if (!ok || !body.success) {
      setItems([]);
      setError((body.message as string) ?? t("loadError"));
      return;
    }
    setError(null);
    setItems(toRows(readInvoicesFromBody(body)));
  }, [t]);

  useEffect(() => {
    let cancelled = false;
    listInvoices().then(({ ok, body }) => {
      if (cancelled) return;
      if (!ok || !body.success) {
        setItems([]);
        setError((body.message as string) ?? t("loadError"));
        return;
      }
      setError(null);
      setItems(toRows(readInvoicesFromBody(body)));
    });
    return () => {
      cancelled = true;
    };
  }, [t]);

  const rowActions = useMemo<DataTableRowAction[]>(
    () => [
      { type: "link", label: t("actions.open"), href: "/invoice/{id}" },
      {
        type: "callback",
        label: t("actions.downloadPdf"),
        callback: (row) => {
          void exportInvoicePdf((row as InvoiceRow).id).catch(() =>
            setError(t("pdfError")),
          );
        },
      },
      {
        type: "callback",
        label: t("actions.markPaid"),
        callback: (row) => {
          const r = row as InvoiceRow;
          if (r.status !== "ISSUED") return;
          void markInvoicePaid(r.id).then(({ ok, body }) => {
            if (!ok || !body.success)
              setError((body.message as string) ?? t("actionError"));
            else void refresh();
          });
        },
      },
      {
        type: "callback",
        label: t("actions.cancel"),
        callback: (row) => {
          const r = row as InvoiceRow;
          if (r.status !== "ISSUED" && r.status !== "PAID") return;
          void cancelInvoice(r.id).then(({ ok, body }) => {
            if (!ok || !body.success)
              setError((body.message as string) ?? t("actionError"));
            else void refresh();
          });
        },
      },
    ],
    [t, refresh],
  );

  return (
    <>
      {error ? <p className="mb-4 text-sm text-destructive">{error}</p> : null}

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
            <DataTableSortableHead name="status">
              {t("columns.status")}
            </DataTableSortableHead>
            <DataTableSortableHead name="quoteId">
              {t("columns.quote")}
            </DataTableSortableHead>
            <DataTableSortableHead name="dueDate">
              {t("columns.dueDate")}
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
                <DataTableCell>{item.number || "—"}</DataTableCell>
                <DataTableCell>
                  <InvoiceStatusBadge status={item.status} />
                </DataTableCell>
                <DataTableCell>{item.quoteId}</DataTableCell>
                <DataTableCell>{item.dueDate || "—"}</DataTableCell>
                <DataTableCell className="tabular-nums">
                  {formatEurosFromCents(item.totalTtc)}
                </DataTableCell>
                <DataTableCell>
                  <DataTableRowActions id={item.id} row={item} />
                </DataTableCell>
              </DataTableRow>
            ))
          )}
        </DataTableBody>
      </DataTable>
    </>
  );
}
