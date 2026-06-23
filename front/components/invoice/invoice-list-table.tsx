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
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { FilterSidebar, FilterSidebarSection } from "@/components/ui/filter-sidebar";
import { SelectCombobox } from "@/components/ui/select-combobox";
import { DateRangePicker } from "@/components/ui/date-range-picker";
import InvoiceStatusBadge from "@/components/invoice/invoice-status-badge";
import InvoiceLifecycleBadge from "@/components/invoice/invoice-lifecycle-badge";
import {
  deleteDraftInvoice,
  listInvoices,
  markInvoicePaid,
  readInvoicesFromBody,
} from "@/lib/services/invoices";
import { listClients } from "@/lib/services/clients";
import { listQuotes } from "@/lib/services/quotes";
import { exportInvoicePdf } from "@/lib/services/export";
import { formatEurosFromCents } from "@/lib/utils";
import type {
  BackendClient,
  BackendInvoiceLifecycleStatus,
  BackendInvoiceStatus,
  BackendInvoiceSummary,
  BackendQuote,
} from "@/types/backend";

const PAGE_SIZE = 20;

const INVOICE_STATUS_ITEMS = [
  { value: "DRAFT", label: "Brouillon" },
  { value: "ISSUED", label: "Émise" },
  { value: "PAID", label: "Payée" },
  { value: "CANCELLED", label: "Annulée" },
];

const LIFECYCLE_STATUS_ITEMS = [
  { value: "NONE", label: "Aucun" },
  { value: "DEPOSITED", label: "Déposée" },
  { value: "RECEIVED", label: "Reçue" },
  { value: "APPROVED", label: "Approuvée" },
  { value: "REJECTED", label: "Rejetée" },
  { value: "COLLECTED", label: "Encaissée" },
];

type InvoiceRow = {
  id: string;
  number: string;
  status: BackendInvoiceStatus;
  lifecycle: BackendInvoiceLifecycleStatus;
  quoteId: string;
  dueDate: string;
  totalTtc: number;
};

function toRows(invoices: BackendInvoiceSummary[]): InvoiceRow[] {
  return invoices.map((i) => ({
    id: i.invoice_id,
    number: i.invoice_number,
    status: i.status,
    lifecycle: i.lifecycle_status,
    quoteId: i.quote_id,
    dueDate: i.due_date,
    totalTtc: i.total_ttc_cents,
  }));
}

function InvoiceListTableInner() {
  const t = useTranslations("invoice.list");
  const tFilters = useTranslations("invoice.list.filters");
  const tCommon = useTranslations("common.filterSidebar");

  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();

  const page = Number(searchParams.get("page") ?? "1");
  const statuses = searchParams.get("statuses") ? searchParams.get("statuses")!.split(",") : [];
  const lifecycleStatuses = searchParams.get("lifecycle_statuses") ? searchParams.get("lifecycle_statuses")!.split(",") : [];
  const issuedFrom = searchParams.get("issued_from") ?? "";
  const issuedTo = searchParams.get("issued_to") ?? "";
  const dueFrom = searchParams.get("due_from") ?? "";
  const dueTo = searchParams.get("due_to") ?? "";
  const clientId = searchParams.get("client_id") ?? "";
  const quoteIdFilter = searchParams.get("quote_id_filter") ?? "";

  const [items, setItems] = useState<InvoiceRow[]>([]);
  const [total, setTotal] = useState(0);
  const [clients, setClients] = useState<BackendClient[]>([]);
  const [quotes, setQuotes] = useState<BackendQuote[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [pendingDeleteId, setPendingDeleteId] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  function pushParams(p: {
    statuses?: string[];
    lifecycleStatuses?: string[];
    issuedFrom?: string;
    issuedTo?: string;
    dueFrom?: string;
    dueTo?: string;
    clientId?: string;
    quoteIdFilter?: string;
    page?: number;
  }) {
    const next = new URLSearchParams();
    const pg = p.page ?? 1;
    const st = p.statuses ?? statuses;
    const lc = p.lifecycleStatuses ?? lifecycleStatuses;
    const iF = p.issuedFrom ?? issuedFrom;
    const iT = p.issuedTo ?? issuedTo;
    const dF = p.dueFrom ?? dueFrom;
    const dT = p.dueTo ?? dueTo;
    const cid = p.clientId ?? clientId;
    const qid = p.quoteIdFilter ?? quoteIdFilter;
    if (pg > 1) next.set("page", String(pg));
    if (st.length > 0) next.set("statuses", st.join(","));
    if (lc.length > 0) next.set("lifecycle_statuses", lc.join(","));
    if (iF) next.set("issued_from", iF);
    if (iT) next.set("issued_to", iT);
    if (dF) next.set("due_from", dF);
    if (dT) next.set("due_to", dT);
    if (cid) next.set("client_id", cid);
    if (qid) next.set("quote_id_filter", qid);
    router.push(`${pathname}?${next.toString()}`);
  }

  useEffect(() => {
    listClients().then(({ ok, body }) => {
      if (ok && Array.isArray(body.clients)) setClients(body.clients as BackendClient[]);
    });
    listQuotes().then(({ ok, body }) => {
      if (ok && Array.isArray(body.quotes)) setQuotes(body.quotes as BackendQuote[]);
    });
  }, []);

  const fetchInvoices = useCallback(async () => {
    const params = new URLSearchParams({ page: String(page), page_size: String(PAGE_SIZE) });
    if (statuses.length > 0) params.set("statuses", statuses.join(","));
    if (lifecycleStatuses.length > 0) params.set("lifecycle_statuses", lifecycleStatuses.join(","));
    if (issuedFrom) params.set("issued_from", issuedFrom);
    if (issuedTo) params.set("issued_to", issuedTo);
    if (dueFrom) params.set("due_from", dueFrom);
    if (dueTo) params.set("due_to", dueTo);
    if (clientId) params.set("client_id", clientId);
    if (quoteIdFilter) params.set("quote_id_filter", quoteIdFilter);

    const { ok, body } = await listInvoices(params.toString());
    if (!ok || !body.success) {
      setError((body.message as string) ?? t("loadError"));
      return;
    }
    setError(null);
    setItems(toRows(readInvoicesFromBody(body)));
    setTotal((body.total ?? 0) as number);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams]);

  useEffect(() => {
    void fetchInvoices();
  }, [fetchInvoices]);

  const clientItems = useMemo(
    () => clients.map((c) => ({
      value: c.client_id,
      label: c.company ? `${c.company} (${c.first_name} ${c.last_name})` : `${c.first_name} ${c.last_name}`,
    })),
    [clients],
  );

  const quoteItems = useMemo(
    () => quotes.map((q) => ({ value: q.quote_id, label: q.name })),
    [quotes],
  );

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

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
            else void fetchInvoices();
          });
        },
      },
      {
        type: "callback",
        label: t("actions.deleteDraft"),
        callback: (row) => {
          const r = row as InvoiceRow;
          if (r.status !== "DRAFT") return;
          setPendingDeleteId(r.id);
        },
      },
    ],
    [t, fetchInvoices],
  );

  async function onConfirmDelete() {
    if (!pendingDeleteId) return;
    setBusy(true);
    const { ok, body } = await deleteDraftInvoice(pendingDeleteId);
    setBusy(false);
    setPendingDeleteId(null);
    if (!ok || !body.success) {
      setError((body.message as string) ?? t("deleteError"));
      return;
    }
    void fetchInvoices();
  }

  const hasFilters = statuses.length > 0 || lifecycleStatuses.length > 0 || issuedFrom || issuedTo || dueFrom || dueTo || clientId || quoteIdFilter;

  return (
    <>
      {error ? <p className="mb-4 text-sm text-destructive">{error}</p> : null}

      <div className="flex items-start gap-4">
        <FilterSidebar
          triggerLabel={tCommon("trigger")}
          title={tCommon("title")}
          activeCount={hasFilters ? 1 : 0}
          onReset={() => pushParams({ statuses: [], lifecycleStatuses: [], issuedFrom: "", issuedTo: "", dueFrom: "", dueTo: "", clientId: "", quoteIdFilter: "", page: 1 })}
          resetLabel={tCommon("reset")}
        >
          <FilterSidebarSection label={tFilters("statusLabel")}>
            <SelectCombobox
              multiple
              items={INVOICE_STATUS_ITEMS}
              value={statuses}
              onValueChange={(v) => pushParams({ statuses: v, page: 1 })}
              placeholder={tFilters("statusPlaceholder")}
              emptyLabel={tFilters("statusEmpty")}
            />
          </FilterSidebarSection>

          <FilterSidebarSection label={tFilters("lifecycleLabel")}>
            <SelectCombobox
              multiple
              items={LIFECYCLE_STATUS_ITEMS}
              value={lifecycleStatuses}
              onValueChange={(v) => pushParams({ lifecycleStatuses: v, page: 1 })}
              placeholder={tFilters("lifecyclePlaceholder")}
              emptyLabel={tFilters("lifecycleEmpty")}
            />
          </FilterSidebarSection>

          <FilterSidebarSection label={tFilters("issuedDateLabel")}>
            <DateRangePicker
              from={issuedFrom}
              to={issuedTo}
              onValueChange={(from, to) => pushParams({ issuedFrom: from, issuedTo: to, page: 1 })}
            />
          </FilterSidebarSection>

          <FilterSidebarSection label={tFilters("dueDateLabel")}>
            <DateRangePicker
              from={dueFrom}
              to={dueTo}
              onValueChange={(from, to) => pushParams({ dueFrom: from, dueTo: to, page: 1 })}
            />
          </FilterSidebarSection>

          <FilterSidebarSection label={tFilters("clientLabel")}>
            <SelectCombobox
              items={clientItems}
              value={clientId}
              onValueChange={(v) => pushParams({ clientId: v, page: 1 })}
              placeholder={tFilters("clientPlaceholder")}
              emptyLabel={tFilters("clientEmpty")}
            />
          </FilterSidebarSection>

          <FilterSidebarSection label={tFilters("quoteLabel")}>
            <SelectCombobox
              items={quoteItems}
              value={quoteIdFilter}
              onValueChange={(v) => pushParams({ quoteIdFilter: v, page: 1 })}
              placeholder={tFilters("quotePlaceholder")}
              emptyLabel={tFilters("quoteEmpty")}
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
                <DataTableSortableHead name="status">
                  {t("columns.status")}
                </DataTableSortableHead>
                <DataTableSortableHead name="lifecycle">
                  {t("columns.lifecycle")}
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
                  <DataTableCell> </DataTableCell>
                </DataTableRow>
              ) : (
                items.map((item) => (
                  <DataTableRow key={item.id}>
                    <DataTableCell>{item.number || "—"}</DataTableCell>
                    <DataTableCell>
                      <InvoiceStatusBadge status={item.status} />
                    </DataTableCell>
                    <DataTableCell>
                      <InvoiceLifecycleBadge status={item.lifecycle} />
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

      <AlertDialog
        open={pendingDeleteId !== null}
        onOpenChange={(open) => {
          if (!open) setPendingDeleteId(null);
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("deleteConfirmTitle")}</AlertDialogTitle>
            <AlertDialogDescription>
              {t("deleteConfirmBody")}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={busy}>
              {t("deleteCancel")}
            </AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={onConfirmDelete}
              disabled={busy}
            >
              {t("deleteConfirmAction")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}

export default function InvoiceListTable() {
  return (
    <Suspense fallback={null}>
      <InvoiceListTableInner />
    </Suspense>
  );
}
