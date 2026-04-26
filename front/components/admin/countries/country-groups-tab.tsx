"use client";

import { useEffect, useState } from "react";
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

function formatMembers(group: CountryGroup): string {
  const list = group.countries ?? [];
  if (list.length === 0) return "—";
  const names = list.map((c) => c.name);
  if (names.length <= 3) return names.join(", ");
  return `${names.slice(0, 3).join(", ")} (+${names.length - 3})`;
}

export default function CountryGroupsTab() {
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
      toast.success("Groupe supprimé.");
      reload();
    } else {
      toast.error(body.message ?? "Une erreur est survenue.");
    }
    setPendingDelete(null);
  }

  const rowActions: DataTableRowAction[] = [
    {
      type: "callback",
      label: "Modifier",
      icon: PencilIcon,
      callback: (row) => openEdit(row as CountryGroup),
    },
    {
      type: "callback",
      label: "Supprimer",
      icon: Trash2Icon,
      callback: (row) => setPendingDelete(row as CountryGroup),
    },
  ];

  return (
    <div className="grid gap-4">
      <div className="flex justify-end">
        <Button type="button" onClick={openCreate}>
          <PlusIcon />
          Nouveau groupe
        </Button>
      </div>

      <DataTable datas={groups} row_actions={rowActions} sortBy="id">
        <DataTableHeader>
          <DataTableRow>
            <DataTableSortableHead name="id">ID</DataTableSortableHead>
            <DataTableSortableHead name="name">Nom</DataTableSortableHead>
            <DataTableHead>Pays membres</DataTableHead>
            <DataTableHead>
              <span className="sr-only">Actions</span>
            </DataTableHead>
          </DataTableRow>
        </DataTableHeader>
        <DataTableBody>
          {groups.length === 0 ? (
            <DataTableRow>
              <DataTableCell className="text-muted-foreground">
                Aucun groupe pour le moment.
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
                  {formatMembers(group)}
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
            <AlertDialogTitle>Supprimer ce groupe ?</AlertDialogTitle>
            <AlertDialogDescription>
              Cette action est irréversible.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Annuler</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={confirmDelete}>
              Supprimer
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
