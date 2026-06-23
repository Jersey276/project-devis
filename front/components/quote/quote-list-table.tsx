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
import { Input } from "@/components/ui/input";
import { FilterSidebar, FilterSidebarSection } from "@/components/ui/filter-sidebar";
import { SelectCombobox } from "@/components/ui/select-combobox";
import { listQuotes, getQuote } from "@/lib/services/quotes";
import { listClients } from "@/lib/services/clients";
import { exportQuotePdf } from "@/lib/services/export";
import {
  createTemplate,
  createTemplateLine,
  deleteTemplate,
} from "@/lib/services/templates";
import { useMode } from "@/lib/mode-context";
import { formatEurosFromCents } from "@/lib/utils";
import {
  type BackendQuote,
  type BackendQuoteLine,
  type BackendClient,
  type QuoteListState,
  quoteListState,
} from "@/types/backend";
import {
  BookmarkIcon,
  CalendarIcon,
  DownloadIcon,
  PencilIcon,
} from "lucide-react";
import { toast } from "sonner";
import SaveTemplateDialog from "@/components/template/save-template-dialog";
import CreateScheduleDialog from "@/components/schedule/create-schedule-dialog";

const PAGE_SIZE = 20;

type QuoteListItem = {
  id: string;
  projectName: string;
  status: QuoteListState;
  totalTtc: number;
};

function QuoteListTableInner() {
  const { isCustomer } = useMode();
  const t = useTranslations("quote.list");
  const tStatus = useTranslations("status.quote");
  const tFilters = useTranslations("quote.list.filters");
  const tCommon = useTranslations("common.filterSidebar");

  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();

  const page = Number(searchParams.get("page") ?? "1");
  const search = searchParams.get("search") ?? "";
  const states = searchParams.get("states") ? searchParams.get("states")!.split(",") : [];
  const clientId = searchParams.get("client_id") ?? "";

  const [items, setItems] = useState<QuoteListItem[]>([]);
  const [total, setTotal] = useState(0);
  const [clients, setClients] = useState<BackendClient[]>([]);
  const [saveTemplateQuoteId, setSaveTemplateQuoteId] = useState<string | null>(null);
  const [saveTemplateDialogOpen, setSaveTemplateDialogOpen] = useState(false);
  const [scheduleQuoteId, setScheduleQuoteId] = useState<string | null>(null);
  const [scheduleDialogOpen, setScheduleDialogOpen] = useState(false);

  function pushParams(p: { search?: string; states?: string[]; clientId?: string; page?: number }) {
    const next = new URLSearchParams();
    const s = p.search ?? search;
    const st = p.states ?? states;
    const cid = p.clientId ?? clientId;
    const pg = p.page ?? 1;
    if (pg > 1) next.set("page", String(pg));
    if (s) next.set("search", s);
    if (st.length > 0) next.set("states", st.join(","));
    if (cid) next.set("client_id", cid);
    router.push(`${pathname}?${next.toString()}`);
  }

  // Load clients once for the combobox
  useEffect(() => {
    if (isCustomer) return;
    listClients().then(({ ok, body }) => {
      if (ok && Array.isArray(body.clients)) setClients(body.clients as BackendClient[]);
    });
  }, [isCustomer]);

  const fetchQuotes = useCallback(async () => {
    if (isCustomer) return;
    const params = new URLSearchParams({ page: String(page), page_size: String(PAGE_SIZE) });
    if (search) params.set("search", search);
    if (states.length > 0) params.set("states", states.join(","));
    if (clientId) params.set("client_id", clientId);

    const { ok, body } = await listQuotes(params.toString());
    if (ok && Array.isArray(body.quotes)) {
      const quotes = body.quotes as BackendQuote[];
      setItems(quotes.map((q) => ({
        id: q.quote_id,
        projectName: q.name,
        status: quoteListState(q),
        totalTtc: q.total_ttc ?? 0,
      })));
      setTotal((body.total ?? 0) as number);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams, isCustomer]);

  useEffect(() => { void fetchQuotes(); }, [fetchQuotes]);

  const saveTemplateDefaultName = useMemo(
    () => saveTemplateQuoteId != null ? (items.find((i) => i.id === saveTemplateQuoteId)?.projectName ?? "") : "",
    [saveTemplateQuoteId, items],
  );

  const handleSaveAsTemplate = useCallback((row: object) => {
    const item = row as QuoteListItem;
    setSaveTemplateQuoteId(item.id);
    setSaveTemplateDialogOpen(true);
  }, []);

  const handleConfirmSaveAsTemplate = useCallback(
    async (name: string): Promise<boolean> => {
      if (!saveTemplateQuoteId) return false;
      const { ok, body } = await getQuote(saveTemplateQuoteId);
      if (!ok || !body.success) { toast.error(t("saveAsTemplateFailedToast")); return false; }
      const lines = (body.lines ?? []) as BackendQuoteLine[];
      const tplRes = await createTemplate({ templateType: "quote_document", targetResource: "quote", name });
      if (!tplRes.ok || !tplRes.body.success) {
        toast.error((tplRes.body.message as string) ?? t("saveAsTemplateFailedToast"));
        return false;
      }
      const templateId = tplRes.body.template_id as string;
      const sorted = [...lines].sort((a, b) => a.position - b.position);
      const lineIdMap = new Map<string, string>();
      for (const [idx, line] of sorted.entries()) {
        const templateParentId = line.data.parent_line_id
          ? (lineIdMap.get(line.data.parent_line_id) ?? line.data.parent_line_id)
          : undefined;
        const lineRes = await createTemplateLine(templateId, {
          type: line.type, name: line.name, quantity: Number(line.quantity),
          unit: line.unit ?? undefined, unitPriceEuros: line.unit_price / 100,
          position: idx, taxId: line.tax_id ?? null,
          data: { ...line.data, parent_line_id: templateParentId },
        });
        if (!lineRes.ok || !lineRes.body.success) {
          await deleteTemplate(templateId);
          toast.error((lineRes.body.message as string) ?? t("saveAsTemplateFailedToast"));
          return false;
        }
        lineIdMap.set(line.line_id, lineRes.body.line_id as string);
      }
      toast.success(t("saveAsTemplateSuccessToast"));
      return true;
    },
    [saveTemplateQuoteId, t],
  );

  const rowActions = useMemo<DataTableRowAction[]>(
    () => [
      { type: "link", label: isCustomer ? t("actions.view") : t("actions.viewEdit"), icon: PencilIcon, href: "/quote/{id}" },
      {
        type: "callback", label: t("actions.exportPdf"), icon: DownloadIcon, hidden: isCustomer,
        callback: (row) => {
          exportQuotePdf((row as { id: string }).id).catch(() => toast.error(t("exportFailedToast")));
        },
      },
      {
        type: "callback", label: "Créer un échéancier", icon: CalendarIcon, hidden: isCustomer,
        callback: (row) => { setScheduleQuoteId((row as { id: string }).id); setScheduleDialogOpen(true); },
      },
      { type: "callback", label: t("actions.saveAsTemplate"), icon: BookmarkIcon, hidden: isCustomer, callback: handleSaveAsTemplate },
    ],
    [isCustomer, t, handleSaveAsTemplate],
  );

  const visibleItems = isCustomer ? [] : items;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const activeFilterCount = (states.length > 0 ? 1 : 0) + (clientId ? 1 : 0);

  const QUOTE_STATE_ITEMS = [
    { value: "draft", label: tStatus("draft") },
    { value: "negociation", label: tStatus("negociation") },
    { value: "validated", label: tStatus("validated") },
    { value: "drop", label: tStatus("drop") },
  ];

  const clientItems = clients.map((c) => ({
    value: c.client_id,
    label: [c.first_name, c.last_name].filter(Boolean).join(" ") || c.company || c.client_id,
  }));

  return (
    <>
      {!isCustomer && (
        <div className="flex flex-wrap items-center gap-2 mb-4">
          <Input
            className="w-full sm:w-64"
            placeholder={tFilters("searchPlaceholder")}
            value={search}
            onChange={(e) => pushParams({ search: e.target.value })}
          />
          <FilterSidebar
            triggerLabel={tCommon("trigger")}
            title={tCommon("title")}
            resetLabel={tCommon("reset")}
            activeCount={activeFilterCount}
            onReset={() => pushParams({ states: [], clientId: "" })}
          >
            <FilterSidebarSection label={tFilters("statusLabel")}>
              <SelectCombobox
                multiple
                items={QUOTE_STATE_ITEMS}
                value={states}
                onValueChange={(vals) => pushParams({ states: vals })}
                placeholder={tFilters("statusPlaceholder")}
                emptyLabel={tFilters("statusEmpty")}
              />
            </FilterSidebarSection>
            <FilterSidebarSection label={tFilters("clientLabel")}>
              <SelectCombobox
                items={clientItems}
                value={clientId}
                onValueChange={(val) => pushParams({ clientId: val })}
                placeholder={tFilters("clientPlaceholder")}
                emptyLabel={tFilters("clientEmpty")}
              />
            </FilterSidebarSection>
          </FilterSidebar>
        </div>
      )}

      <DataTable datas={visibleItems} sortBy="id" sortDirection="asc" row_actions={rowActions}>
        <DataTableHeader>
          <DataTableRow>
            <DataTableSortableHead name="id">{t("columns.id")}</DataTableSortableHead>
            <DataTableSortableHead name="projectName">{t("columns.project")}</DataTableSortableHead>
            <DataTableSortableHead name="status">{t("columns.status")}</DataTableSortableHead>
            <DataTableSortableHead name="totalTtc">{t("columns.totalTtc")}</DataTableSortableHead>
            <DataTableHead>{t("columns.actions")}</DataTableHead>
          </DataTableRow>
        </DataTableHeader>
        <DataTableBody>
          {visibleItems.length === 0 ? (
            <DataTableRow>
              {[...Array(5)].map((_, i) => (
                <DataTableCell key={i} className={i === 0 ? "text-muted-foreground" : ""}>{i === 0 ? t("empty") : " "}</DataTableCell>
              ))}
            </DataTableRow>
          ) : (
            visibleItems.map((quote) => (
              <DataTableRow key={quote.id}>
                <DataTableCell>{quote.id}</DataTableCell>
                <DataTableCell>{quote.projectName}</DataTableCell>
                <DataTableCell>{tStatus(quote.status)}</DataTableCell>
                <DataTableCell className="tabular-nums">{formatEurosFromCents(quote.totalTtc)}</DataTableCell>
                <DataTableCell><DataTableRowActions id={quote.id} row={quote} /></DataTableCell>
              </DataTableRow>
            ))
          )}
        </DataTableBody>
      </DataTable>

      {!isCustomer && total > 0 && (
        <div className="flex flex-wrap items-center justify-between gap-2 mt-4 text-sm text-muted-foreground">
          <span>{total} devis</span>
          <div className="flex gap-2">
            <button className="rounded border px-3 py-1 disabled:opacity-40" disabled={page <= 1} onClick={() => pushParams({ page: page - 1 })}>←</button>
            <span>{page} / {totalPages}</span>
            <button className="rounded border px-3 py-1 disabled:opacity-40" disabled={page >= totalPages} onClick={() => pushParams({ page: page + 1 })}>→</button>
          </div>
        </div>
      )}

      <SaveTemplateDialog
        open={saveTemplateDialogOpen}
        onOpenChange={setSaveTemplateDialogOpen}
        defaultName={saveTemplateDefaultName}
        onSave={handleConfirmSaveAsTemplate}
      />
      <CreateScheduleDialog
        open={scheduleDialogOpen}
        onOpenChange={setScheduleDialogOpen}
        initialQuoteId={scheduleQuoteId ?? undefined}
        lockQuote
      />
    </>
  );
}

export default function QuoteListTable() {
  return (
    <Suspense fallback={null}>
      <QuoteListTableInner />
    </Suspense>
  );
}
