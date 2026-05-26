"use client";

import { useTranslations } from "next-intl";
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
import { EyeIcon, TrashIcon } from "lucide-react";
import { toast } from "sonner";
import { archiveClient } from "@/lib/services/clients";
import type { BackendClient } from "@/types/backend";

type ClientsTableProps = {
  data: BackendClient[];
  onArchived?: () => void;
};

export function ClientsTable({ data, onArchived }: ClientsTableProps) {
  const t = useTranslations("client.list");
  const tCommon = useTranslations("common");
  const row_actions: DataTableRowAction[] = [
    {
      type: "link",
      label: t("actions.view"),
      href: "/clients/{id}",
      icon: EyeIcon,
    },
    {
      type: "callback",
      label: tCommon("actions.delete"),
      icon: TrashIcon,
      callback: async (row) => {
        const client = row as BackendClient;
        const { ok, body } = await archiveClient(client.client_id);
        if (ok && body.success) {
          toast.success(t("deleteSuccessToast"));
          onArchived?.();
        } else {
          toast.error(
            (body.message as string) ?? t("deleteFailedToast"),
          );
        }
      },
    },
  ];

  return (
    <DataTable
      datas={data}
      sortBy="client_id"
      sortDirection="asc"
      row_actions={row_actions}
    >
      <DataTableHeader>
        <DataTableRow>
          <DataTableSortableHead name="first_name">
            {t("columns.firstName")}
          </DataTableSortableHead>
          <DataTableSortableHead name="last_name">{t("columns.lastName")}</DataTableSortableHead>
          <DataTableSortableHead name="email">{t("columns.email")}</DataTableSortableHead>
          <DataTableSortableHead name="company">{t("columns.company")}</DataTableSortableHead>
          <DataTableHead>{t("columns.actions")}</DataTableHead>
        </DataTableRow>
      </DataTableHeader>
      <DataTableBody>
        {data.map((client) => (
          <DataTableRow key={client.client_id}>
            <DataTableCell>{client.first_name}</DataTableCell>
            <DataTableCell>{client.last_name}</DataTableCell>
            <DataTableCell>{client.email}</DataTableCell>
            <DataTableCell>{client.company}</DataTableCell>
            <DataTableCell>
              <DataTableRowActions id={client.client_id} row={client} />
            </DataTableCell>
          </DataTableRow>
        ))}
      </DataTableBody>
    </DataTable>
  );
}
