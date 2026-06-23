"use client";

import { Suspense, useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams, useRouter, usePathname } from "next/navigation";
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
import { Button } from "@/components/ui/button";
import { FilterSidebar, FilterSidebarSection } from "@/components/ui/filter-sidebar";
import { SelectCombobox } from "@/components/ui/select-combobox";
import { DateRangePicker } from "@/components/ui/date-range-picker";
import { listSchedules } from "@/lib/services/schedules";
import type { BackendScheduleSummary } from "@/types/backend";
import CreateScheduleDialog from "@/components/schedule/create-schedule-dialog";
import ScheduleStatusSelect from "@/components/schedule/schedule-status-select";

const PAGE_SIZE = 20;

const SCHEDULE_STATUS_ITEMS = [
  { value: "DRAFT", label: "Brouillon" },
  { value: "NEGOCIATE", label: "En négociation" },
  { value: "DENIED", label: "Refusé" },
  { value: "VALID", label: "Validé" },
];

type ScheduleRow = {
  id: string;
  quoteId: string;
  name: string;
  status: string;
  startMonth: string;
  durationMonths: number;
};

function toRows(schedules: BackendScheduleSummary[]): ScheduleRow[] {
  return schedules.map((s) => ({
    id: s.schedule_id,
    quoteId: s.quote_id,
    name: s.name,
    status: s.status,
    startMonth: s.start_month,
    durationMonths: s.duration_months,
  }));
}

function ScheduleListTableInner() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();

  const page = Number(searchParams.get("page") ?? "1");
  const statuses = searchParams.get("statuses") ? searchParams.get("statuses")!.split(",") : [];
  const startFrom = searchParams.get("start_from") ?? "";
  const startTo = searchParams.get("start_to") ?? "";

  const [items, setItems] = useState<ScheduleRow[]>([]);
  const [total, setTotal] = useState(0);
  const [open, setOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function pushParams(p: { statuses?: string[]; startFrom?: string; startTo?: string; page?: number }) {
    const next = new URLSearchParams();
    const st = p.statuses ?? statuses;
    const sf = p.startFrom ?? startFrom;
    const stt = p.startTo ?? startTo;
    const pg = p.page ?? 1;
    if (pg > 1) next.set("page", String(pg));
    if (st.length > 0) next.set("statuses", st.join(","));
    if (sf) next.set("start_from", sf);
    if (stt) next.set("start_to", stt);
    router.push(`${pathname}?${next.toString()}`);
  }

  const refreshSchedules = useCallback(async () => {
    const params = new URLSearchParams({ page: String(page), page_size: String(PAGE_SIZE) });
    if (statuses.length > 0) params.set("statuses", statuses.join(","));
    if (startFrom) params.set("start_from", startFrom);
    if (startTo) params.set("start_to", startTo);

    const { ok, body } = await listSchedules(params.toString());
    if (!ok || !body.success || !Array.isArray(body.schedules)) {
      setItems([]);
      setError((body.message as string) ?? "Impossible de charger les échéanciers.");
      return;
    }
    setError(null);
    setItems(toRows(body.schedules as BackendScheduleSummary[]));
    setTotal((body.total ?? 0) as number);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams]);

  useEffect(() => { void refreshSchedules(); }, [refreshSchedules]);

  const rowActions = useMemo<DataTableRowAction[]>(() => [{ type: "link", label: "Ouvrir", href: "/schedule/{id}" }], []);

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const activeFilterCount = (statuses.length > 0 ? 1 : 0) + (startFrom || startTo ? 1 : 0);

  return (
    <>
      <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
        <FilterSidebar
          triggerLabel="Filtres"
          title="Filtres"
          resetLabel="Réinitialiser les filtres"
          activeCount={activeFilterCount}
          onReset={() => pushParams({ statuses: [], startFrom: "", startTo: "" })}
        >
          <FilterSidebarSection label="Statut">
            <SelectCombobox
              multiple
              items={SCHEDULE_STATUS_ITEMS}
              value={statuses}
              onValueChange={(vals) => pushParams({ statuses: vals })}
              placeholder="Sélectionner des statuts…"
              emptyLabel="Aucun statut trouvé."
            />
          </FilterSidebarSection>
          <FilterSidebarSection label="Mois de départ">
            <DateRangePicker
              from={startFrom}
              to={startTo}
              onValueChange={(from, to) => pushParams({ startFrom: from, startTo: to, page: 1 })}
            />
          </FilterSidebarSection>
        </FilterSidebar>
        <Button type="button" onClick={() => setOpen(true)}>
          Nouvel échéancier
        </Button>
      </div>

      {error ? <p className="mb-4 text-sm text-destructive">{error}</p> : null}

      <DataTable datas={items} sortBy="startMonth" sortDirection="desc" row_actions={rowActions}>
        <DataTableHeader>
          <DataTableRow>
            <DataTableSortableHead name="id">ID</DataTableSortableHead>
            <DataTableSortableHead name="name">Nom</DataTableSortableHead>
            <DataTableSortableHead name="quoteId">Devis</DataTableSortableHead>
            <DataTableSortableHead name="status">Statut</DataTableSortableHead>
            <DataTableSortableHead name="startMonth">Début</DataTableSortableHead>
            <DataTableSortableHead name="durationMonths">Durée (mois)</DataTableSortableHead>
            <DataTableHead>Actions</DataTableHead>
          </DataTableRow>
        </DataTableHeader>
        <DataTableBody>
          {items.length === 0 ? (
            <DataTableRow>
              {[...Array(7)].map((_, i) => (
                <DataTableCell key={i} className={i === 0 ? "text-muted-foreground" : ""}>{i === 0 ? "Aucun échéancier." : " "}</DataTableCell>
              ))}
            </DataTableRow>
          ) : (
            items.map((item) => (
              <DataTableRow key={item.id}>
                <DataTableCell>{item.id}</DataTableCell>
                <DataTableCell>{item.name}</DataTableCell>
                <DataTableCell>{item.quoteId}</DataTableCell>
                <DataTableCell>
                  <ScheduleStatusSelect
                    scheduleId={item.id}
                    value={item.status as BackendScheduleSummary["status"]}
                    className="w-44"
                    onUpdated={refreshSchedules}
                    onError={setError}
                  />
                </DataTableCell>
                <DataTableCell>{item.startMonth}</DataTableCell>
                <DataTableCell>{item.durationMonths}</DataTableCell>
                <DataTableCell><DataTableRowActions id={item.id} row={item} /></DataTableCell>
              </DataTableRow>
            ))
          )}
        </DataTableBody>
      </DataTable>

      {total > 0 && (
        <div className="flex flex-wrap items-center justify-between gap-2 mt-4 text-sm text-muted-foreground">
          <span>{total} échéancier{total > 1 ? "s" : ""}</span>
          <div className="flex gap-2">
            <button className="rounded border px-3 py-1 disabled:opacity-40" disabled={page <= 1} onClick={() => pushParams({ page: page - 1 })}>←</button>
            <span>{page} / {totalPages}</span>
            <button className="rounded border px-3 py-1 disabled:opacity-40" disabled={page >= totalPages} onClick={() => pushParams({ page: page + 1 })}>→</button>
          </div>
        </div>
      )}

      <CreateScheduleDialog open={open} onOpenChange={setOpen} onCreated={refreshSchedules} />
    </>
  );
}

export default function ScheduleListTable() {
  return (
    <Suspense fallback={null}>
      <ScheduleListTableInner />
    </Suspense>
  );
}
