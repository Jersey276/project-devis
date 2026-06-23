"use client";

import { useCallback, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { PencilIcon } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import ProjectStatusBadge from "@/components/project/project-status-badge";
import ProjectCharts from "@/components/project/project-charts";
import ProjectQuotesTable from "@/components/project/project-quotes-table";
import EditProjectDialog from "@/components/project/edit-project-dialog";
import { getProjectDetail } from "@/lib/services/projects";
import { listClients } from "@/lib/services/clients";
import { clientName } from "@/lib/project-utils";
import type { BackendClient, BackendProjectDetail } from "@/types/backend";

export default function ProjectDetail({ projectId }: { projectId: string }) {
  const t = useTranslations("project");

  const [detail, setDetail] = useState<BackendProjectDetail | null>(null);
  const [clients, setClients] = useState<BackendClient[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editOpen, setEditOpen] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  const load = useCallback(() => {
    const controller = new AbortController();
    setLoading(true);
    getProjectDetail(projectId, controller.signal).then(({ ok, body }) => {
      setLoading(false);
      if (!ok) { setError(body?.message ?? t("errors.load")); return; }
      setDetail(body as unknown as BackendProjectDetail);
    }).catch(() => {});
    return () => controller.abort();
  }, [projectId, refreshKey, t]);

  useEffect(() => {
    const cleanup = load();
    return cleanup;
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [refreshKey, projectId]);

  useEffect(() => {
    listClients().then(({ ok, body }) => {
      if (ok && Array.isArray(body.clients)) setClients(body.clients);
    });
  }, []);

  if (loading) {
    return (
      <div className="flex flex-col gap-4">
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-48 w-full" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (error || !detail) {
    return <p className="text-sm text-destructive">{error ?? t("errors.load")}</p>;
  }

  const project = detail.project;
  const name = clientName(project.client_id, clients, "");

  return (
    <div className="flex flex-col gap-6">
      {/* Header */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between gap-4">
          <div className="flex flex-col gap-1">
            <CardTitle className="text-xl">{project.name}</CardTitle>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              {name && <span>{name}</span>}
              <ProjectStatusBadge status={project.status} />
            </div>
          </div>
          <Button variant="outline" size="sm" onClick={() => setEditOpen(true)}>
            <PencilIcon className="size-3.5" />
            Modifier
          </Button>
        </CardHeader>
      </Card>

      {/* Charts */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Vue d&apos;ensemble</CardTitle>
        </CardHeader>
        <CardContent>
          <ProjectCharts
            quotes={detail.quotes}
            totalHtCents={detail.total_ht_cents}
            collectedHtCents={detail.collected_ht_cents}
          />
        </CardContent>
      </Card>

      {/* Quotes table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Devis du projet</CardTitle>
        </CardHeader>
        <CardContent>
          <ProjectQuotesTable
            projectId={projectId}
            quotes={detail.quotes}
            onChanged={() => setRefreshKey((k) => k + 1)}
          />
        </CardContent>
      </Card>

      {editOpen && (
        <EditProjectDialog
          project={project}
          open={editOpen}
          onOpenChange={setEditOpen}
          onUpdated={() => setRefreshKey((k) => k + 1)}
        />
      )}
    </div>
  );
}
