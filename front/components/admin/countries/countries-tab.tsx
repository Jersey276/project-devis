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
import CountryDialog from "./country-dialog";
import { type Country } from "@/components/address/address-form";

export default function CountriesTab() {
  const [countries, setCountries] = useState<Country[]>([]);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<Country | null>(null);
  const [pendingDelete, setPendingDelete] = useState<Country | null>(null);
  const [reloadKey, setReloadKey] = useState(0);

  const reload = () => setReloadKey((k) => k + 1);

  useEffect(() => {
    let cancelled = false;
    apiFetch("/api/users/countries").then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && Array.isArray(body.countries)) {
        setCountries(body.countries as Country[]);
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

  function openEdit(country: Country) {
    setEditing(country);
    setDialogOpen(true);
  }

  async function confirmDelete() {
    if (!pendingDelete) return;
    const { ok, body } = await apiFetch(
      `/api/users/countries/${pendingDelete.id}`,
      { method: "DELETE" },
    );
    if (ok && body.success) {
      toast.success("Pays supprimé.");
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
      callback: (row) => openEdit(row as Country),
    },
    {
      type: "callback",
      label: "Supprimer",
      icon: Trash2Icon,
      callback: (row) => setPendingDelete(row as Country),
    },
  ];

  return (
    <div className="grid gap-4">
      <div className="flex justify-end">
        <Button type="button" onClick={openCreate}>
          <PlusIcon />
          Nouveau pays
        </Button>
      </div>

      <DataTable datas={countries} row_actions={rowActions} sortBy="id">
        <DataTableHeader>
          <DataTableRow>
            <DataTableSortableHead name="id">ID</DataTableSortableHead>
            <DataTableSortableHead name="code">Code</DataTableSortableHead>
            <DataTableSortableHead name="name">Nom</DataTableSortableHead>
            <DataTableHead>
              <span className="sr-only">Actions</span>
            </DataTableHead>
          </DataTableRow>
        </DataTableHeader>
        <DataTableBody>
          {countries.length === 0 ? (
            <DataTableRow>
              <DataTableCell className="text-muted-foreground">
                Aucun pays pour le moment.
              </DataTableCell>
              <DataTableCell> </DataTableCell>
              <DataTableCell> </DataTableCell>
              <DataTableCell> </DataTableCell>
            </DataTableRow>
          ) : (
            countries.map((country) => (
              <DataTableRow key={country.id}>
                <DataTableCell>{country.id}</DataTableCell>
                <DataTableCell>{country.code}</DataTableCell>
                <DataTableCell>{country.name}</DataTableCell>
                <DataTableCell className="w-12 text-right">
                  <DataTableRowActions id={country.id} row={country} />
                </DataTableCell>
              </DataTableRow>
            ))
          )}
        </DataTableBody>
      </DataTable>

      <CountryDialog
        key={editing?.id ?? "new"}
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        country={editing}
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
            <AlertDialogTitle>Supprimer ce pays ?</AlertDialogTitle>
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
