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
import { Badge } from "@/components/ui/badge";
import {
  listCreditNotes,
  readCreditNotesFromBody,
} from "@/lib/services/invoices";
import { exportCreditNotePdf } from "@/lib/services/export";
import { formatEurosFromCents } from "@/lib/utils";
import type { BackendCreditNoteSummary } from "@/types/backend";

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

export default function CreditNoteListTable() {
  const t = useTranslations("creditNote.list");
  const [items, setItems] = useState<CreditNoteRow[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    listCreditNotes().then(({ ok, body }) => {
      if (cancelled) return;
      if (!ok || !body.success) {
        setItems([]);
        setError((body.message as string) ?? t("loadError"));
        return;
      }
      setError(null);
      setItems(toRows(readCreditNotesFromBody(body)));
    });
    return () => {
      cancelled = true;
    };
  }, [t]);

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
    </>
  );
}
