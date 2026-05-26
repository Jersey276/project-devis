"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
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
import {
  DataTable,
  DataTableBody,
  DataTableCell,
  DataTableHead,
  DataTableHeader,
  DataTableRow,
  DataTableRowActions,
  DataTableSortableHead,
  type DataTableRowAction,
} from "@/components/custom/data-table";
import { PencilIcon, PlusIcon, Trash2Icon } from "lucide-react";
import { apiFetch } from "@/lib/api";
import { toast } from "sonner";
import CountryGroupDialog from "./country-group-dialog";
import { type CountryGroup } from "@/components/admin/types";

function formatMembers(group: CountryGroup, dash: string): string {
  const list = group.countries ?? [];
  if (list.length === 0) return dash;
  const names = list.map((c) => c.name);
  if (names.length <= 3) return names.join(", ");
  return `${names.slice(0, 3).join(", ")} (+${names.length - 3})`;
}

export default function CountryGroupsTab() {
  const t = useTranslations("admin.countryGroups");
  const tCommon = useTranslations("common");
  const [groups, setGroups] = useState<CountryGroup[]>([]);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<CountryGroup | null>(null);
  const [pendingDelete, setPendingDelete] = useState<CountryGroup | null>(null);
  const [reloadKey, setReloadKey] = useState(0);

  const reload = () => setReloadKey((k) => k + 1);

  useEffect(() => {
    let cancelled = false;
    apiFetch("/api/users/country-groups").then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && Array.isArray(body.country_groups)) {
        setGroups(body.country_groups as CountryGroup[]);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [reloadKey]);

  function openCreate() {
    setEditing(null);
    setDialogOpen(true);
  }

  function openEdit(group: CountryGroup) {
    setEditing(group);
    setDialogOpen(true);
  }

  async function confirmDelete() {
    if (!pendingDelete) return;
    const { ok, body } = await apiFetch(
      `/api/users/country-groups/${pendingDelete.id}`,
      { method: "DELETE" },
    );
    if (ok && body.success) {
      toast.success(t("deleteSuccessToast"));
      reload();
    } else {
      toast.error(body.message ?? tCommon("errors.generic"));
    }
    setPendingDelete(null);
  }

  const rowActions: DataTableRowAction[] = [
    {
      type: "callback",
      label: tCommon("actions.edit"),
      icon: PencilIcon,
      callback: (row) => openEdit(row as CountryGroup),
    },
    {
      type: "callback",
      label: tCommon("actions.delete"),
      icon: Trash2Icon,
      callback: (row) => setPendingDelete(row as CountryGroup),
    },
  ];

  return (
    <div className="grid gap-4">
      <div className="flex justify-end">
        <Button type="button" onClick={openCreate}>
          <PlusIcon />
          {t("newButton")}
        </Button>
      </div>

      <DataTable datas={groups} row_actions={rowActions} sortBy="id">
        <DataTableHeader>
          <DataTableRow>
            <DataTableSortableHead name="id">{t("columns.id")}</DataTableSortableHead>
            <DataTableSortableHead name="name">{t("columns.name")}</DataTableSortableHead>
            <DataTableHead>{t("columns.members")}</DataTableHead>
            <DataTableHead>
              <span className="sr-only">{t("actionsLabel")}</span>
            </DataTableHead>
          </DataTableRow>
        </DataTableHeader>
        <DataTableBody>
          {groups.length === 0 ? (
            <DataTableRow>
              <DataTableCell className="text-muted-foreground">
                {t("empty")}
              </DataTableCell>
              <DataTableCell> </DataTableCell>
              <DataTableCell> </DataTableCell>
              <DataTableCell> </DataTableCell>
            </DataTableRow>
          ) : (
            groups.map((group) => (
              <DataTableRow key={group.id}>
                <DataTableCell>{group.id}</DataTableCell>
                <DataTableCell>{group.name}</DataTableCell>
                <DataTableCell className="text-muted-foreground">
                  {formatMembers(group, t("membersEmptyDash"))}
                </DataTableCell>
                <DataTableCell className="w-12 text-right">
                  <DataTableRowActions id={group.id} row={group} />
                </DataTableCell>
              </DataTableRow>
            ))
          )}
        </DataTableBody>
      </DataTable>

      <CountryGroupDialog
        key={editing?.id ?? "new"}
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        group={editing}
        onSaved={reload}
      />

      <AlertDialog
        open={pendingDelete !== null}
        onOpenChange={(open) => {
          if (!open) setPendingDelete(null);
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("deleteDialog.title")}</AlertDialogTitle>
            <AlertDialogDescription>
              {t("deleteDialog.description")}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{tCommon("actions.cancel")}</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={confirmDelete}>
              {tCommon("actions.delete")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
