"use client";

import { Suspense, useEffect, useMemo, useState } from "react";
import { useSearchParams, useRouter, usePathname } from "next/navigation";
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
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { FilterSidebar, FilterSidebarSection } from "@/components/ui/filter-sidebar";
import { SelectCombobox } from "@/components/ui/select-combobox";
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
import { BanIcon, EyeIcon, PencilIcon } from "lucide-react";
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

function UsersTableInner() {
  const t = useTranslations("admin.users");
  const tCommon = useTranslations("common");
  const tFilters = useTranslations("admin.users.filters");
  const tFilterSidebar = useTranslations("common.filterSidebar");

  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();

  const search = searchParams.get("search") ?? "";
  const roles = useMemo(
    () => (searchParams.get("roles") ? searchParams.get("roles")!.split(",") : []),
    [searchParams],
  );
  const statuses = useMemo(
    () => (searchParams.get("statuses") ? searchParams.get("statuses")!.split(",") : []),
    [searchParams],
  );

  function pushParams(newSearch: string, newRoles: string[], newStatuses: string[]) {
    const p = new URLSearchParams();
    if (newSearch) p.set("search", newSearch);
    if (newRoles.length > 0) p.set("roles", newRoles.join(","));
    if (newStatuses.length > 0) p.set("statuses", newStatuses.join(","));
    router.push(`${pathname}?${p.toString()}`);
  }

  const { key: reloadKey, reload } = useReloadKey();
  const [users, setUsers] = useState<AdminUserAccount[]>([]);
  const [editing, setEditing] = useState<AdminUserAccount | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [pendingSuspend, setPendingSuspend] = useState<AdminUserAccount | null>(null);

  useEffect(() => {
    let cancelled = false;
    const params = new URLSearchParams();
    if (search) params.set("search", search);
    if (roles.length > 0) params.set("roles", roles.join(","));
    if (statuses.length > 0) params.set("statuses", statuses.join(","));
    listAdminUsers(params.toString() || undefined).then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && Array.isArray(body.users)) {
        setUsers(body.users as AdminUserAccount[]);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [reloadKey, search, roles, statuses]);

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
      type: "link",
      label: tCommon("actions.view"),
      icon: EyeIcon,
      href: "/users/{id}",
    },
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

  const ROLE_ITEMS = [
    { value: "user", label: t("roleUser") },
    { value: "admin", label: t("roleAdmin") },
  ];

  const STATUS_ITEMS = [
    { value: "active", label: tFilters("statusActive") },
    { value: "suspended", label: tFilters("statusSuspended") },
  ];

  const activeFilterCount = (roles.length > 0 ? 1 : 0) + (statuses.length > 0 ? 1 : 0);

  return (
    <div className="grid gap-4">
      <div className="flex flex-wrap items-center gap-2">
        <Input
          className="w-full sm:w-64"
          placeholder={tFilters("searchPlaceholder")}
          value={search}
          onChange={(e) => pushParams(e.target.value, roles, statuses)}
        />
        <FilterSidebar
          triggerLabel={tFilterSidebar("trigger")}
          title={tFilterSidebar("title")}
          resetLabel={tFilterSidebar("reset")}
          activeCount={activeFilterCount}
          onReset={() => pushParams(search, [], [])}
        >
          <FilterSidebarSection label={tFilters("roleLabel")}>
            <SelectCombobox
              multiple
              items={ROLE_ITEMS}
              value={roles}
              onValueChange={(vals) => pushParams(search, vals, statuses)}
              placeholder={tFilters("rolePlaceholder")}
              emptyLabel={tFilters("roleEmpty")}
            />
          </FilterSidebarSection>
          <FilterSidebarSection label={tFilters("statusLabel")}>
            <SelectCombobox
              multiple
              items={STATUS_ITEMS}
              value={statuses}
              onValueChange={(vals) => pushParams(search, roles, vals)}
              placeholder={tFilters("statusPlaceholder")}
              emptyLabel={tFilters("statusEmpty")}
            />
          </FilterSidebarSection>
        </FilterSidebar>
      </div>

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
        <DataTableBodyRows<AdminUserAccount>
          emptyColSpan={6}
          empty={<span className="text-muted-foreground">{t("empty")}</span>}
          render={(user) => (
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
          )}
        />
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

export default function UsersTable() {
  return (
    <Suspense fallback={null}>
      <UsersTableInner />
    </Suspense>
  );
}
