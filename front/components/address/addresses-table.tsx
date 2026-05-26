"use client";

import { useCallback, useEffect, useState } from "react";
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
  type DataTableRowAction,
} from "@/components/custom/data-table";
import { PencilIcon, PlusIcon, Trash2Icon } from "lucide-react";
import AddressDialog, {
  type ExistingAddress,
} from "@/components/address/address-dialog";
import { apiFetch } from "@/lib/api";
import { type Country } from "@/components/address/address-form";
import {
  archiveAddress,
  buildOwner,
  listAddresses,
} from "@/lib/services/addresses";
import type { BackendAddress } from "@/types/backend";
import { toast } from "sonner";

function formatAddress(address: BackendAddress, countries: Country[]): string {
  const country = countries.find((c) => c.id === address.country_id);
  const lines = [
    address.street,
    address.additional_street ? address.additional_street : null,
    `${address.zip_code} ${address.city}`,
    country?.name ?? "",
  ].filter(Boolean);
  return lines.join(", ");
}

type AddressesTableProps = {
  ownerType: "user" | "client";
  ownerId: string;
};

export default function AddressesTable({
  ownerType,
  ownerId,
}: AddressesTableProps) {
  const t = useTranslations("address.list");
  const tCommon = useTranslations("common");
  const [addresses, setAddresses] = useState<BackendAddress[]>([]);
  const [countries, setCountries] = useState<Country[]>([]);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editing, setEditing] = useState<ExistingAddress | null>(null);
  const [pendingDelete, setPendingDelete] = useState<BackendAddress | null>(
    null,
  );

  const reload = useCallback(() => {
    listAddresses(buildOwner(ownerType, ownerId)).then(({ ok, body }) => {
      if (ok && Array.isArray(body.addresses)) {
        setAddresses(body.addresses as BackendAddress[]);
      }
    });
  }, [ownerType, ownerId]);

  useEffect(() => {
    reload();
  }, [reload]);

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
  }, []);

  function openCreate() {
    setEditing(null);
    setDrawerOpen(true);
  }

  function openEdit(address: BackendAddress) {
    setEditing({
      id: address.id,
      name: address.name,
      street: address.street,
      additional_street: address.additional_street ?? "",
      city: address.city,
      zip_code: address.zip_code,
      country_id: address.country_id,
    });
    setDrawerOpen(true);
  }

  async function confirmDelete() {
    if (!pendingDelete) return;
    const { ok, body } = await archiveAddress(
      buildOwner(ownerType, ownerId),
      pendingDelete.id,
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
      callback: (row) => openEdit(row as BackendAddress),
    },
    {
      type: "callback",
      label: tCommon("actions.delete"),
      icon: Trash2Icon,
      callback: (row) => setPendingDelete(row as BackendAddress),
    },
  ];

  return (
    <div className="grid gap-4">
      <div className="flex justify-end">
        <Button type="button" onClick={openCreate}>
          <PlusIcon />
          {t("addButton")}
        </Button>
      </div>

      <DataTable datas={addresses} row_actions={rowActions} sortBy="">
        <DataTableHeader>
          <DataTableRow>
            <DataTableHead>{t("columns.address")}</DataTableHead>
            <DataTableHead>
              <span className="sr-only">{t("actionsLabel")}</span>
            </DataTableHead>
          </DataTableRow>
        </DataTableHeader>
        <DataTableBody>
          {addresses.length === 0 ? (
            <DataTableRow>
              <DataTableCell className="text-muted-foreground">
                {t("empty")}
              </DataTableCell>
              <DataTableCell> </DataTableCell>
            </DataTableRow>
          ) : (
            addresses.map((address) => (
              <DataTableRow key={address.id}>
                <DataTableCell>
                  <div className="font-medium">{address.name}</div>
                  <div className="text-muted-foreground text-sm">
                    {formatAddress(address, countries)}
                  </div>
                </DataTableCell>
                <DataTableCell className="w-12 text-right">
                  <DataTableRowActions id={address.id} row={address} />
                </DataTableCell>
              </DataTableRow>
            ))
          )}
        </DataTableBody>
      </DataTable>

      <AddressDialog
        ownerType={ownerType}
        ownerId={ownerId}
        open={drawerOpen}
        onOpenChange={setDrawerOpen}
        address={editing}
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
            <AlertDialogAction
              variant="destructive"
              onClick={confirmDelete}
            >
              {tCommon("actions.delete")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
