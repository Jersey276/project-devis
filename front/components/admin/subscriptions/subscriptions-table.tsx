"use client";

import { useEffect, useState } from "react";
import { useReloadKey } from "@/hooks/use-reload-key";
import { useTranslations } from "next-intl";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
import { PencilIcon } from "lucide-react";
import { toast } from "sonner";
import {
  listAdminSubscriptions,
  listPlans,
  assignPlan,
} from "@/lib/services/subscriptions";
import type {
  BackendSubscription,
  BackendPlan,
  SubscriptionTier,
} from "@/types/backend";

function formatDate(value: string | null | undefined): string {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return new Intl.DateTimeFormat("fr-FR", { dateStyle: "medium" }).format(date);
}

export default function SubscriptionsTable() {
  const t = useTranslations("admin.subscriptions");
  const tCommon = useTranslations("common");

  const { key: reloadKey, reload } = useReloadKey();
  const [subscriptions, setSubscriptions] = useState<BackendSubscription[]>([]);
  const [plans, setPlans] = useState<BackendPlan[]>([]);
  const [editing, setEditing] = useState<BackendSubscription | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedPlanId, setSelectedPlanId] = useState<string>("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    let cancelled = false;

    listAdminSubscriptions().then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && Array.isArray(body.subscriptions)) {
        setSubscriptions(body.subscriptions as BackendSubscription[]);
      }
    });

    listPlans().then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && Array.isArray(body.plans)) {
        setPlans(body.plans as BackendPlan[]);
      }
    });

    return () => {
      cancelled = true;
    };
  }, [reloadKey]);

  function openEdit(sub: BackendSubscription) {
    setEditing(sub);
    const matchingPlan = plans.find((p) => p.tier === sub.tier);
    setSelectedPlanId(matchingPlan ? String(matchingPlan.plan_id) : "");
    setDialogOpen(true);
  }

  async function confirmAssign() {
    if (!editing || !selectedPlanId) return;

    setSubmitting(true);
    try {
      const { ok, body } = await assignPlan(
        editing.user_id,
        Number(selectedPlanId),
      );
      if (ok && body.success) {
        toast.success(t("changePlanDialog.successToast"));
        reload();
        setDialogOpen(false);
      } else {
        toast.error(body.message ?? tCommon("errors.generic"));
      }
    } catch {
      toast.error(tCommon("errors.generic"));
    } finally {
      setSubmitting(false);
    }
  }

  const tierVariant = (tier: SubscriptionTier) => {
    if (tier === "enterprise") return "default" as const;
    if (tier === "pro") return "secondary" as const;
    return "outline" as const;
  };

  const rowActions: DataTableRowAction[] = [
    {
      type: "callback",
      label: t("actions.changePlan"),
      icon: PencilIcon,
      callback: (row) => openEdit(row as BackendSubscription),
    },
  ];

  return (
    <div className="grid gap-4">
      <DataTable
        datas={subscriptions}
        row_actions={rowActions}
        sortBy="updated_at"
      >
        <DataTableHeader>
          <DataTableRow>
            <DataTableSortableHead name="user_id">
              {t("columns.userId")}
            </DataTableSortableHead>
            <DataTableSortableHead name="tier">
              {t("columns.plan")}
            </DataTableSortableHead>
            <DataTableSortableHead name="status">
              {t("columns.status")}
            </DataTableSortableHead>
            <DataTableSortableHead name="updated_at">
              {t("columns.updatedAt")}
            </DataTableSortableHead>
            <DataTableHead>
              <span className="sr-only">{t("actionsLabel")}</span>
            </DataTableHead>
          </DataTableRow>
        </DataTableHeader>
        <DataTableBody>
          {subscriptions.length === 0 ? (
            <DataTableRow>
              <DataTableCell className="text-muted-foreground">
                {t("empty")}
              </DataTableCell>
              <DataTableCell> </DataTableCell>
              <DataTableCell> </DataTableCell>
              <DataTableCell> </DataTableCell>
              <DataTableCell> </DataTableCell>
            </DataTableRow>
          ) : (
            subscriptions.map((sub) => (
              <DataTableRow key={sub.subscription_id || sub.user_id}>
                <DataTableCell className="font-mono text-xs">
                  {sub.user_id}
                </DataTableCell>
                <DataTableCell>
                  <Badge variant={tierVariant(sub.tier)}>
                    {t(`tiers.${sub.tier}`)}
                  </Badge>
                </DataTableCell>
                <DataTableCell>{sub.status}</DataTableCell>
                <DataTableCell className="text-muted-foreground">
                  {formatDate(sub.updated_at)}
                </DataTableCell>
                <DataTableCell className="w-12 text-right">
                  <DataTableRowActions
                    id={sub.subscription_id || sub.user_id}
                    row={sub}
                  />
                </DataTableCell>
              </DataTableRow>
            ))
          )}
        </DataTableBody>
      </DataTable>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>{t("changePlanDialog.title")}</DialogTitle>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="assign_plan_select">
                {t("changePlanDialog.planLabel")}
              </FieldLabel>
              <Select value={selectedPlanId} onValueChange={setSelectedPlanId}>
                <SelectTrigger id="assign_plan_select" className="w-full">
                  <SelectValue
                    placeholder={t("changePlanDialog.planPlaceholder")}
                  />
                </SelectTrigger>
                <SelectContent>
                  {plans.map((p) => (
                    <SelectItem key={p.plan_id} value={String(p.plan_id)}>
                      {p.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDialogOpen(false)}
              disabled={submitting}
            >
              {tCommon("actions.cancel")}
            </Button>
            <Button
              onClick={confirmAssign}
              disabled={submitting || !selectedPlanId}
            >
              {submitting
                ? tCommon("actions.saving")
                : t("changePlanDialog.confirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
