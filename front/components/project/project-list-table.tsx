"use client";

import { Suspense, useEffect, useMemo, useState } from "react";
import { useSearchParams, useRouter, usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import Link from "next/link";
import {
  DataTable,
  DataTableBodyRows,
  DataTableCell,
  DataTableHead,
  DataTableHeader,
  DataTableRow,
  type DataTableRowAction,
  DataTableRowActions,
  DataTableSortableHead,
} from "@/components/custom/data-table";
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
import { Input } from "@/components/ui/input";
import { FilterSidebar, FilterSidebarSection } from "@/components/ui/filter-sidebar";
import { SelectCombobox } from "@/components/ui/select-combobox";
import ProjectStatusBadge from "@/components/project/project-status-badge";
import EditProjectDialog, { type EditableProject } from "@/components/project/edit-project-dialog";
import { listProjects, deleteProject } from "@/lib/services/projects";
import { listClients } from "@/lib/services/clients";
import { clientName, PROJECT_STATUS_ITEMS } from "@/lib/project-utils";
import { toast } from "sonner";
import type { BackendClient, ProjectStatus } from "@/types/backend";

const PAGE_SIZE = 20;

type ProjectRow = {
  id: string;
  name: string;
  clientId: string;
  status: ProjectStatus;
  quoteCount: number;
  createdAt: string;
};

function toRows(projects: BackendProject[]): ProjectRow[] {
  return projects.map((p) => ({
    id: p.project_id,
    name: p.name,
    clientId: p.client_id,
    status: p.status,
    quoteCount: p.quote_count,
    createdAt: p.created_at,
  }));
}


function ProjectListTableInner() {
  const t = useTranslations("project");
  const tCommon = useTranslations("common.filterSidebar");

  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();

  const page = Number(searchParams.get("page") ?? "1");
  const search = searchParams.get("search") ?? "";
  const status = searchParams.get("status") ?? "";
  const sortBy = searchParams.get("sort_by") ?? "created_at";
  const sortDirection = (searchParams.get("sort_direction") ?? "desc") as "asc" | "desc";

  const [items, setItems] = useState<ProjectRow[]>([]);
  const [total, setTotal] = useState(0);
  const [clients, setClients] = useState<BackendClient[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [pendingDeleteId, setPendingDeleteId] = useState<string | null>(null);
  const [editProject, setEditProject] = useState<EditableProject | null>(null);
  const [busy, setBusy] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  function pushParams(p: {
    page?: number;
    search?: string;
    status?: string;
    sort_by?: string;
    sort_direction?: string;
  }) {
    const params = new URLSearchParams(searchParams.toString());
    if (p.page !== undefined) params.set("page", String(p.page));
    if (p.search !== undefined) { p.search ? params.set("search", p.search) : params.delete("search"); }
    if (p.status !== undefined) { p.status ? params.set("status", p.status) : params.delete("status"); }
    if (p.sort_by !== undefined) params.set("sort_by", p.sort_by);
    if (p.sort_direction !== undefined) params.set("sort_direction", p.sort_direction);
    router.push(`${pathname}?${params.toString()}`);
  }

  // Load clients once for name resolution
  useEffect(() => {
    listClients().then(({ ok, body }) => {
      if (ok && Array.isArray(body.clients)) setClients(body.clients);
    });
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    const params = new URLSearchParams({
      page: String(page),
      page_size: String(PAGE_SIZE),
      sort_by: sortBy,
      sort_direction: sortDirection,
    });
    if (search) params.set("search", search);
    if (status) params.set("status", status);

    listProjects(params.toString(), controller.signal).then(({ ok, body }) => {
      if (!ok) { setError(body?.message ?? "Erreur de chargement."); return; }
      setError(null);
      setItems(toRows(body.projects ?? []));
      setTotal(body.total ?? 0);
    }).catch(() => {});

    return () => controller.abort();
  }, [page, search, status, sortBy, sortDirection, refreshKey]);

  const activeFilterCount = useMemo(() => [search, status].filter(Boolean).length, [search, status]);

  async function handleDelete() {
    if (!pendingDeleteId) return;
    setBusy(true);
    const { ok, body } = await deleteProject(pendingDeleteId);
    setBusy(false);
    setPendingDeleteId(null);
    if (!ok) { toast.error(body?.message ?? "La suppression a échoué."); return; }
    toast.success("Projet supprimé.");
    setRefreshKey((k) => k + 1);
  }

  const rowActions = useMemo<DataTableRowAction[]>(() => [
    { type: "link", label: "Voir", href: "/projects/{id}" },
    {
      type: "callback",
      label: "Modifier",
      callback: (row) => {
        const p = row as ProjectRow;
        setEditProject({ project_id: p.id, name: p.name, client_id: p.clientId, status: p.status });
      },
    },
    {
      type: "callback",
      label: "Supprimer",
      callback: (row) => setPendingDeleteId((row as ProjectRow).id),
    },
  ], []);

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  const filters = (
    <FilterSidebar activeCount={activeFilterCount}>
      <FilterSidebarSection label={tCommon("search")}>
        <Input
          value={search}
          onChange={(e) => pushParams({ search: e.target.value, page: 1 })}
          placeholder="Rechercher…"
        />
      </FilterSidebarSection>
      <FilterSidebarSection label={t("list.columns.status")}>
        <SelectCombobox
          items={PROJECT_STATUS_ITEMS}
          value={status}
          onValueChange={(v) => pushParams({ status: v, page: 1 })}
          placeholder="Tous"
        />
      </FilterSidebarSection>
    </FilterSidebar>
  );

  return (
    <>
      {error && <p className="text-sm text-destructive mb-2">{error}</p>}
      <DataTable
        datas={items as object[]}
        row_actions={rowActions}
        filters={filters}
        sortBy={sortBy}
        sortDirection={sortDirection}
        onSortChange={(col, dir) => pushParams({ sort_by: col, sort_direction: dir, page: 1 })}
      >
        <DataTableHeader>
          <DataTableRow>
            <DataTableSortableHead column="name">{t("list.columns.name")}</DataTableSortableHead>
            <DataTableHead>{t("list.columns.client")}</DataTableHead>
            <DataTableHead>{t("list.columns.status")}</DataTableHead>
            <DataTableHead>{t("list.columns.quoteCount")}</DataTableHead>
            <DataTableHead>{t("list.columns.actions")}</DataTableHead>
          </DataTableRow>
        </DataTableHeader>
        <DataTableBodyRows
          render={(row: ProjectRow) => (
            <DataTableRow key={row.id}>
              <DataTableCell>
                <Link href={`/projects/${row.id}`} className="font-medium hover:underline">
                  {row.name}
                </Link>
              </DataTableCell>
              <DataTableCell>{clientName(row.clientId, clients)}</DataTableCell>
              <DataTableCell>
                <ProjectStatusBadge status={row.status} />
              </DataTableCell>
              <DataTableCell>{row.quoteCount}</DataTableCell>
              <DataTableCell>
                <DataTableRowActions id={row.id} row={row as object} />
              </DataTableCell>
            </DataTableRow>
          )}
        />
      </DataTable>

      {total > 0 ? (
        <div className="mt-4 flex items-center justify-between gap-2 text-sm text-muted-foreground">
          <span>{total} projet{total > 1 ? "s" : ""}</span>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => pushParams({ page: page - 1 })}>
              Précédent
            </Button>
            <span>{page} / {totalPages}</span>
            <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => pushParams({ page: page + 1 })}>
              Suivant
            </Button>
          </div>
        </div>
      ) : (
        !error && <p className="text-sm text-muted-foreground mt-4">{t("list.empty")}</p>
      )}

      <AlertDialog open={!!pendingDeleteId} onOpenChange={(v) => { if (!v) setPendingDeleteId(null); }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Supprimer le projet ?</AlertDialogTitle>
            <AlertDialogDescription>
              Cette action est irréversible.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={busy}>Annuler</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete} disabled={busy}>
              Supprimer
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {editProject && (
        <EditProjectDialog
          project={editProject}
          open={!!editProject}
          onOpenChange={(v) => { if (!v) setEditProject(null); }}
          onUpdated={() => { setEditProject(null); setRefreshKey((k) => k + 1); }}
        />
      )}
    </>
  );
}

export default function ProjectListTable() {
  return (
    <Suspense>
      <ProjectListTableInner />
    </Suspense>
  );
}
