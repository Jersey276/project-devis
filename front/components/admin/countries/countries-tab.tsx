"use client";

import { useState } from "react";
import { useCountries } from "@/hooks/use-countries";
import { useReloadKey } from "@/hooks/use-reload-key";
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

import CountryDialog from "./country-dialog";
import { type Country } from "@/components/address/address-form";

export default function CountriesTab() {
  const t = useTranslations("admin.countries");
  const tCommon = useTranslations("common");
  const { key: reloadKey, reload } = useReloadKey();
  const countries = useCountries(reloadKey);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<Country | null>(null);
  const [pendingDelete, setPendingDelete] = useState<Country | null>(null);

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
      callback: (row) => openEdit(row as Country),
    },
    {
      type: "callback",
      label: tCommon("actions.delete"),
      icon: Trash2Icon,
      callback: (row) => setPendingDelete(row as Country),
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

      <DataTable datas={countries} row_actions={rowActions} sortBy="id">
        <DataTableHeader>
          <DataTableRow>
            <DataTableSortableHead name="id">
              {t("columns.id")}
            </DataTableSortableHead>
            <DataTableSortableHead name="code">
              {t("columns.code")}
            </DataTableSortableHead>
            <DataTableSortableHead name="name">
              {t("columns.name")}
            </DataTableSortableHead>
            <DataTableHead>
              <span className="sr-only">{t("actionsLabel")}</span>
            </DataTableHead>
          </DataTableRow>
        </DataTableHeader>
        <DataTableBody>
          {countries.length === 0 ? (
            <DataTableRow>
              <DataTableCell className="text-muted-foreground">
                {t("empty")}
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
