"use client";

import { useEffect, useState } from "react";
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
  DataTableBodyRows,
  DataTableCell,
  DataTableHead,
  DataTableHeader,
  DataTableRow,
  DataTableRowActions,
  DataTableSortableHead,
  type DataTableRowAction,
} from "@/components/custom/data-table";
import { Trash2Icon, PencilIcon, PlusIcon } from "lucide-react";
import { toast } from "sonner";
import { listFees, archiveFee } from "@/lib/services/fees";
import FeeDialog from "./fee-dialog";
import type { BackendFee } from "@/types/backend";

export default function FeesTable() {
  const t = useTranslations("fees.list");
  const tCategories = useTranslations("fees.categories");
  const tCommon = useTranslations("common");
  const { key: reloadKey, reload } = useReloadKey();
  const [fees, setFees] = useState<BackendFee[]>([]);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<BackendFee | null>(null);
  const [pendingDelete, setPendingDelete] = useState<BackendFee | null>(null);

  useEffect(() => {
    let cancelled = false;
    listFees().then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && Array.isArray(body.fees)) {
        setFees(body.fees as BackendFee[]);
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

  function openEdit(fee: BackendFee) {
    setEditing(fee);
    setDialogOpen(true);
  }

  async function confirmDelete() {
    if (!pendingDelete) return;
    const { ok, body } = await archiveFee(pendingDelete.fee_id);
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
      callback: (row) => openEdit(row as BackendFee),
    },
    {
      type: "callback",
      label: tCommon("actions.delete"),
      icon: Trash2Icon,
      callback: (row) => setPendingDelete(row as BackendFee),
    },
  ];

  function formatPrice(cents: number): string {
    return t("priceValue", { amount: (cents / 100).toFixed(2) });
  }

  return (
    <div className="grid gap-4">
      <div className="flex justify-end">
        <Button type="button" onClick={openCreate}>
          <PlusIcon />
          {t("newButton")}
        </Button>
      </div>

      <DataTable datas={fees} row_actions={rowActions} sortBy="name">
        <DataTableHeader>
          <DataTableRow>
            <DataTableSortableHead name="name">
              {t("columns.name")}
            </DataTableSortableHead>
            <DataTableSortableHead name="category">
              {t("columns.category")}
            </DataTableSortableHead>
            <DataTableHead>{t("columns.unit")}</DataTableHead>
            <DataTableSortableHead name="unit_price">
              {t("columns.price")}
            </DataTableSortableHead>
            <DataTableHead>
              <span className="sr-only">{t("actionsLabel")}</span>
            </DataTableHead>
          </DataTableRow>
        </DataTableHeader>
        <DataTableBodyRows<BackendFee>
          emptyColSpan={5}
          empty={<span className="text-muted-foreground">{t("empty")}</span>}
          render={(fee) => (
            <DataTableRow key={fee.fee_id}>
              <DataTableCell>{fee.name}</DataTableCell>
              <DataTableCell>{tCategories(fee.category)}</DataTableCell>
              <DataTableCell>{fee.unit || "—"}</DataTableCell>
              <DataTableCell>{formatPrice(fee.unit_price)}</DataTableCell>
              <DataTableCell className="w-12 text-right">
                <DataTableRowActions id={fee.fee_id} row={fee} />
              </DataTableCell>
            </DataTableRow>
          )}
        />
      </DataTable>

      <FeeDialog
        key={editing?.fee_id ?? "new"}
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        fee={editing}
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
