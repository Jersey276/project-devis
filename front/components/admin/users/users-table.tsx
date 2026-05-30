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
import { Badge } from "@/components/ui/badge";
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
import { BanIcon, PencilIcon } from "lucide-react";
import { toast } from "sonner";
import { listAdminUsers, suspendAdminUser } from "@/lib/services/admin-users";
import { type AdminUserAccount } from "@/components/admin/types";
import UserEditDialog from "./user-edit-dialog";

function formatLastLogin(value: string | null, fallback: string): string {
  if (!value) return fallback;
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return fallback;
  return new Intl.DateTimeFormat("fr-FR", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

export default function UsersTable() {
  const t = useTranslations("admin.users");
  const tCommon = useTranslations("common");

  const [users, setUsers] = useState<AdminUserAccount[]>([]);
  const [editing, setEditing] = useState<AdminUserAccount | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [pendingSuspend, setPendingSuspend] = useState<AdminUserAccount | null>(
    null,
  );
  const [reloadKey, setReloadKey] = useState(0);

  const reload = () => setReloadKey((k) => k + 1);

  useEffect(() => {
    let cancelled = false;
    listAdminUsers().then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && Array.isArray(body.users)) {
        setUsers(body.users as AdminUserAccount[]);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [reloadKey]);

  function openEdit(user: AdminUserAccount) {
    setEditing(user);
    setDialogOpen(true);
  }

  async function confirmSuspend() {
    if (!pendingSuspend) return;

    const { ok, body } = await suspendAdminUser(pendingSuspend.user_id);
    if (ok && body.success) {
      toast.success(t("suspendSuccessToast"));
      reload();
    } else {
      toast.error(body.message ?? tCommon("errors.generic"));
    }
    setPendingSuspend(null);
  }

  const rowActions: DataTableRowAction[] = [
    {
      type: "callback",
      label: tCommon("actions.edit"),
      icon: PencilIcon,
      callback: (row) => openEdit(row as AdminUserAccount),
    },
    {
      type: "callback",
      label: t("actions.suspend"),
      icon: BanIcon,
      callback: (row) => setPendingSuspend(row as AdminUserAccount),
    },
  ];

  return (
    <div className="grid gap-4">
      <DataTable datas={users} row_actions={rowActions} sortBy="last_name">
        <DataTableHeader>
          <DataTableRow>
            <DataTableSortableHead name="last_name">
              {t("columns.lastName")}
            </DataTableSortableHead>
            <DataTableSortableHead name="first_name">
              {t("columns.firstName")}
            </DataTableSortableHead>
            <DataTableSortableHead name="email">
              {t("columns.email")}
            </DataTableSortableHead>
            <DataTableSortableHead name="role">
              {t("columns.role")}
            </DataTableSortableHead>
            <DataTableSortableHead name="last_login_at">
              {t("columns.lastLogin")}
            </DataTableSortableHead>
            <DataTableHead>
              <span className="sr-only">{t("actionsLabel")}</span>
            </DataTableHead>
          </DataTableRow>
        </DataTableHeader>
        <DataTableBody>
          {users.length === 0 ? (
            <DataTableRow>
              <DataTableCell className="text-muted-foreground">
                {t("empty")}
              </DataTableCell>
              <DataTableCell> </DataTableCell>
              <DataTableCell> </DataTableCell>
              <DataTableCell> </DataTableCell>
              <DataTableCell> </DataTableCell>
              <DataTableCell> </DataTableCell>
            </DataTableRow>
          ) : (
            users.map((user) => (
              <DataTableRow key={user.user_id}>
                <DataTableCell>{user.last_name || "-"}</DataTableCell>
                <DataTableCell>{user.first_name || "-"}</DataTableCell>
                <DataTableCell>{user.email}</DataTableCell>
                <DataTableCell>
                  <Badge
                    variant={user.role === "admin" ? "secondary" : "outline"}
                  >
                    {user.role === "admin" ? t("roleAdmin") : t("roleUser")}
                  </Badge>
                </DataTableCell>
                <DataTableCell className="text-muted-foreground">
                  {formatLastLogin(user.last_login_at, t("neverLoggedIn"))}
                </DataTableCell>
                <DataTableCell className="w-12 text-right">
                  <DataTableRowActions id={user.user_id} row={user} />
                </DataTableCell>
              </DataTableRow>
            ))
          )}
        </DataTableBody>
      </DataTable>

      <UserEditDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        user={editing}
        onSaved={reload}
      />

      <AlertDialog
        open={pendingSuspend !== null}
        onOpenChange={(open) => {
          if (!open) setPendingSuspend(null);
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("suspendDialog.title")}</AlertDialogTitle>
            <AlertDialogDescription>
              {t("suspendDialog.description")}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{tCommon("actions.cancel")}</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={confirmSuspend}>
              {t("actions.suspend")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
