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
import { listSchedules } from "@/lib/services/schedules";
import type { BackendScheduleSummary } from "@/types/backend";

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

export default function ScheduleListTable() {
  const [items, setItems] = useState<ScheduleRow[]>([]);

  const rowActions = useMemo<DataTableRowAction[]>(
    () => [
      {
        type: "link",
        label: "Ouvrir",
        href: "/schedule/{id}",
      },
    ],
    [],
  );

  useEffect(() => {
    let cancelled = false;
    listSchedules().then(({ ok, body }) => {
      if (cancelled) return;
      if (!ok || !body.success || !Array.isArray(body.schedules)) {
        setItems([]);
        return;
      }
      setItems(toRows(body.schedules as BackendScheduleSummary[]));
    });
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <DataTable
      datas={items}
      sortBy="startMonth"
      sortDirection="desc"
      row_actions={rowActions}
    >
      <DataTableHeader>
        <DataTableRow>
          <DataTableSortableHead name="id">ID</DataTableSortableHead>
          <DataTableSortableHead name="name">Nom</DataTableSortableHead>
          <DataTableSortableHead name="quoteId">Devis</DataTableSortableHead>
          <DataTableSortableHead name="status">Statut</DataTableSortableHead>
          <DataTableSortableHead name="startMonth">Début</DataTableSortableHead>
          <DataTableSortableHead name="durationMonths">
            Durée (mois)
          </DataTableSortableHead>
          <DataTableHead>Actions</DataTableHead>
        </DataTableRow>
      </DataTableHeader>
      <DataTableBody>
        {items.length === 0 ? (
          <DataTableRow>
            <DataTableCell className="text-muted-foreground">
              Aucun échéancier.
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
              <DataTableCell>{item.id}</DataTableCell>
              <DataTableCell>{item.name}</DataTableCell>
              <DataTableCell>{item.quoteId}</DataTableCell>
              <DataTableCell>{item.status}</DataTableCell>
              <DataTableCell>{item.startMonth}</DataTableCell>
              <DataTableCell>{item.durationMonths}</DataTableCell>
              <DataTableCell>
                <DataTableRowActions id={item.id} row={item} />
              </DataTableCell>
            </DataTableRow>
          ))
        )}
      </DataTableBody>
    </DataTable>
  );
}
