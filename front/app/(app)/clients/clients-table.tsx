"use client";

import {
  DataTable,
  DataTableBody,
  DataTableCell,
  DataTableHeader,
  DataTableHead,
  DataTableRow,
  DataTableRowAction,
  DataTableRowActions,
  DataTableSortableHead,
} from "@/components/custom/data-table";
import { TrashIcon } from "lucide-react";

type Client = {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
};

type ClientsTableProps = {
  data: Client[];
};

export function ClientsTable({ data }: ClientsTableProps) {
  const row_actions: DataTableRowAction[] = [
    {
      type: "callback",
      label: "Supprimer",
      icon: TrashIcon,
      callback: (row) => {
        console.log("Supprimer client:", row);
      },
    },
  ];

  return (
    <DataTable
      datas={data}
      sortBy="id"
      sortDirection="asc"
      row_actions={row_actions}
    >
      <DataTableHeader>
        <DataTableRow>
          <DataTableSortableHead name="id">ID</DataTableSortableHead>
          <DataTableSortableHead name="first_name">
            Prénom
          </DataTableSortableHead>
          <DataTableSortableHead name="last_name">Nom</DataTableSortableHead>
          <DataTableSortableHead name="email">Email</DataTableSortableHead>
          <DataTableHead>Actions</DataTableHead>
        </DataTableRow>
      </DataTableHeader>
      <DataTableBody>
        {data.map((client, index) => (
          <DataTableRow key={index}>
            <DataTableCell>{client.id}</DataTableCell>
            <DataTableCell>{client.first_name}</DataTableCell>
            <DataTableCell>{client.last_name}</DataTableCell>
            <DataTableCell>{client.email}</DataTableCell>
            <DataTableCell>
              <DataTableRowActions id={client.id} row={client} />
            </DataTableCell>
          </DataTableRow>
        ))}
      </DataTableBody>
    </DataTable>
  );
}
