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
import { ArchiveIcon, CheckIcon, PencilIcon, PlusIcon } from "lucide-react";
import { apiFetch } from "@/lib/api";
import { toast } from "sonner";
import TaxDialog from "./tax-dialog";
import { type CountryGroup, type Tax } from "@/components/admin/types";

export default function TaxesTable() {
  const [taxes, setTaxes] = useState<Tax[]>([]);
  const [groups, setGroups] = useState<CountryGroup[]>([]);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<Tax | null>(null);
  const [pendingDelete, setPendingDelete] = useState<Tax | null>(null);
  const [reloadKey, setReloadKey] = useState(0);

  const reload = () => setReloadKey((k) => k + 1);

  useEffect(() => {
    let cancelled = false;
    apiFetch("/api/users/taxes").then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && Array.isArray(body.taxes)) {
        setTaxes(body.taxes as Tax[]);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [reloadKey]);

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
  }, []);

  function openCreate() {
    setEditing(null);
    setDialogOpen(true);
  }

  function openEdit(tax: Tax) {
    setEditing(tax);
    setDialogOpen(true);
  }

  async function confirmDelete() {
    if (!pendingDelete) return;
    const { ok, body } = await apiFetch(
      `/api/users/taxes/${pendingDelete.id}`,
      { method: "DELETE" },
    );
    if (ok && body.success) {
      toast.success("Taxe retirée.");
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
      callback: (row) => openEdit(row as Tax),
    },
    {
      type: "callback",
      label: "Retirer",
      icon: ArchiveIcon,
      callback: (row) => setPendingDelete(row as Tax),
    },
  ];

  function groupName(id: number): string {
    return groups.find((g) => g.id === id)?.name ?? `#${id}`;
  }

  return (
    <div className="grid gap-4">
      <div className="flex justify-end">
        <Button type="button" onClick={openCreate}>
          <PlusIcon />
          Nouvelle taxe
        </Button>
      </div>

      <DataTable datas={taxes} row_actions={rowActions} sortBy="id">
        <DataTableHeader>
          <DataTableRow>
            <DataTableSortableHead name="id">ID</DataTableSortableHead>
            <DataTableSortableHead name="name">Nom</DataTableSortableHead>
            <DataTableSortableHead name="rate">Taux</DataTableSortableHead>
            <DataTableHead>Groupe</DataTableHead>
            <DataTableHead>Défaut</DataTableHead>
            <DataTableHead>
              <span className="sr-only">Actions</span>
            </DataTableHead>
          </DataTableRow>
        </DataTableHeader>
        <DataTableBody>
          {taxes.length === 0 ? (
            <DataTableRow>
              <DataTableCell className="text-muted-foreground">
                Aucune taxe pour le moment.
              </DataTableCell>
              <DataTableCell> </DataTableCell>
              <DataTableCell> </DataTableCell>
              <DataTableCell> </DataTableCell>
              <DataTableCell> </DataTableCell>
              <DataTableCell> </DataTableCell>
            </DataTableRow>
          ) : (
            taxes.map((tax) => (
              <DataTableRow key={tax.id}>
                <DataTableCell>{tax.id}</DataTableCell>
                <DataTableCell>{tax.name}</DataTableCell>
                <DataTableCell>{tax.rate} %</DataTableCell>
                <DataTableCell>{groupName(tax.country_group_id)}</DataTableCell>
                <DataTableCell>
                  {tax.is_default ? (
                    <CheckIcon
                      className="size-4 text-emerald-600"
                      aria-label="Taxe par défaut"
                    />
                  ) : null}
                </DataTableCell>
                <DataTableCell className="w-12 text-right">
                  <DataTableRowActions id={tax.id} row={tax} />
                </DataTableCell>
              </DataTableRow>
            ))
          )}
        </DataTableBody>
      </DataTable>

      <TaxDialog
        key={editing?.id ?? "new"}
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        tax={editing}
        groups={groups}
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
            <AlertDialogTitle>Retirer cette taxe ?</AlertDialogTitle>
            <AlertDialogDescription>
              La taxe restera consultable sur les devis qui la référencent.
              Aucun nouveau devis ne pourra l&apos;utiliser.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Annuler</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={confirmDelete}>
              Retirer
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
