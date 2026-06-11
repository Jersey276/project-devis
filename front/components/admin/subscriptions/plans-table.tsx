"use client";

import { useEffect, useState } from "react";
import { useReloadKey } from "@/hooks/use-reload-key";
import { useTranslations } from "next-intl";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Field,
  FieldGroup,
  FieldDescription,
  FieldLabel,
} from "@/components/ui/field";
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
import { listAllPlans, updatePlan } from "@/lib/services/subscriptions";
import type { BackendPlan, SubscriptionTier } from "@/types/backend";

const BILLING_CYCLES = ["monthly", "yearly", "none"] as const;
const PLAN_FEATURE_KEYS = ["max_schedules", "max_templates"] as const;

function parsePlanFeatures(
  features: BackendPlan["features"],
): Record<string, number> {
  if (typeof features === "string") {
    try {
      return JSON.parse(features);
    } catch {
      return {};
    }
  }
  return features ?? {};
}

function centsToEuros(cents: number): string {
  return (cents / 100).toFixed(2);
}

function eurosToCents(euros: string): number {
  return Math.round(parseFloat(euros || "0") * 100);
}

export default function PlansTable() {
  const t = useTranslations("admin.plans");
  const tCommon = useTranslations("common");

  const { key: reloadKey, reload } = useReloadKey();
  const [plans, setPlans] = useState<BackendPlan[]>([]);
  const [editing, setEditing] = useState<BackendPlan | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const [formName, setFormName] = useState("");
  const [formPriceEuros, setFormPriceEuros] = useState("");
  const [formBillingCycle, setFormBillingCycle] = useState<string>("none");
  const [formStripePriceId, setFormStripePriceId] = useState("");
  const [formFeatures, setFormFeatures] = useState<Record<string, number>>({});

  useEffect(() => {
    let cancelled = false;
    listAllPlans().then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && Array.isArray(body.plans)) {
        setPlans(body.plans as BackendPlan[]);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [reloadKey]);

  function openEdit(plan: BackendPlan) {
    setEditing(plan);
    setFormName(plan.name);
    setFormPriceEuros(centsToEuros(plan.price_cents));
    setFormBillingCycle(plan.billing_cycle);
    setFormStripePriceId(plan.stripe_price_id ?? "");
    const parsed = parsePlanFeatures(plan.features);
    const initialFeatures: Record<string, number> = {};
    for (const key of PLAN_FEATURE_KEYS) {
      initialFeatures[key] = parsed[key] ?? 0;
    }
    setFormFeatures(initialFeatures);
    setDialogOpen(true);
  }

  async function confirmUpdate() {
    if (!editing) return;

    setSubmitting(true);
    try {
      const { ok, body } = await updatePlan(editing.plan_id, {
        name: formName,
        price_cents: eurosToCents(formPriceEuros),
        billing_cycle: formBillingCycle,
        stripe_price_id: formStripePriceId,
        features: JSON.stringify(formFeatures),
      });
      if (ok && body.success) {
        toast.success(t("editDialog.successToast"));
        reload();
        setDialogOpen(false);
      } else {
        toast.error(body.message ?? tCommon("errors.generic"));
      }
    } catch {
      toast.error(t("editDialog.errorToast"));
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
      label: t("actions.edit"),
      icon: PencilIcon,
      callback: (row) => openEdit(row as BackendPlan),
    },
  ];

  return (
    <div className="grid gap-4">
      <DataTable datas={plans} row_actions={rowActions} sortBy="tier">
        <DataTableHeader>
          <DataTableRow>
            <DataTableSortableHead name="name">
              {t("columns.name")}
            </DataTableSortableHead>
            <DataTableSortableHead name="tier">
              {t("columns.tier")}
            </DataTableSortableHead>
            <DataTableSortableHead name="price_cents">
              {t("columns.price")}
            </DataTableSortableHead>
            <DataTableSortableHead name="billing_cycle">
              {t("columns.billingCycle")}
            </DataTableSortableHead>
            <DataTableHead>{t("columns.stripePriceId")}</DataTableHead>
            <DataTableHead>
              <span className="sr-only">{t("actionsLabel")}</span>
            </DataTableHead>
          </DataTableRow>
        </DataTableHeader>
        <DataTableBody>
          {plans.map((plan) => (
            <DataTableRow key={plan.plan_id}>
              <DataTableCell className="font-medium">{plan.name}</DataTableCell>
              <DataTableCell>
                <Badge variant={tierVariant(plan.tier)}>{plan.tier}</Badge>
              </DataTableCell>
              <DataTableCell>
                {plan.price_cents === 0
                  ? "—"
                  : new Intl.NumberFormat("fr-FR", {
                      style: "currency",
                      currency: "EUR",
                    }).format(plan.price_cents / 100)}
              </DataTableCell>
              <DataTableCell>
                {t(`billingCycle.${plan.billing_cycle}`)}
              </DataTableCell>
              <DataTableCell className="font-mono text-xs text-muted-foreground">
                {plan.stripe_price_id || "—"}
              </DataTableCell>
              <DataTableCell className="w-12 text-right">
                <DataTableRowActions id={String(plan.plan_id)} row={plan} />
              </DataTableCell>
            </DataTableRow>
          ))}
        </DataTableBody>
      </DataTable>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{t("editDialog.title")}</DialogTitle>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="plan_name">
                {t("editDialog.nameLabel")}
              </FieldLabel>
              <Input
                id="plan_name"
                value={formName}
                onChange={(e) => setFormName(e.target.value)}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="plan_price">
                {t("editDialog.priceLabelEuros")}
              </FieldLabel>
              <Input
                id="plan_price"
                type="number"
                min="0"
                step="0.01"
                value={formPriceEuros}
                onChange={(e) => setFormPriceEuros(e.target.value)}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="plan_billing_cycle">
                {t("editDialog.billingCycleLabel")}
              </FieldLabel>
              <Select
                value={formBillingCycle}
                onValueChange={setFormBillingCycle}
              >
                <SelectTrigger id="plan_billing_cycle" className="w-full">
                  <SelectValue
                    placeholder={t("editDialog.billingCyclePlaceholder")}
                  />
                </SelectTrigger>
                <SelectContent>
                  {BILLING_CYCLES.map((cycle) => (
                    <SelectItem key={cycle} value={cycle}>
                      {t(`billingCycle.${cycle}`)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
            <Field>
              <FieldLabel htmlFor="plan_stripe_price_id">
                {t("editDialog.stripePriceIdLabel")}
              </FieldLabel>
              <Input
                id="plan_stripe_price_id"
                placeholder="price_xxx"
                value={formStripePriceId}
                onChange={(e) => setFormStripePriceId(e.target.value)}
              />
              <FieldDescription>
                {t("editDialog.stripePriceIdHint")}
              </FieldDescription>
            </Field>
            <Field>
              <FieldLabel>{t("editDialog.featuresLabel")}</FieldLabel>
              <FieldGroup>
                {PLAN_FEATURE_KEYS.map((key) => {
                  const isUnlimited = formFeatures[key] === -1;
                  return (
                    <div key={key} className="flex items-center gap-3">
                      <span className="flex-1 text-sm">
                        {t(`editDialog.features.${key}`)}
                      </span>
                      <label className="flex items-center gap-1.5 text-sm cursor-pointer select-none">
                        <Checkbox
                          checked={isUnlimited}
                          onCheckedChange={(checked) =>
                            setFormFeatures((prev) => ({
                              ...prev,
                              [key]: checked ? -1 : 0,
                            }))
                          }
                        />
                        {t("editDialog.features.unlimited")}
                      </label>
                      <Input
                        type="number"
                        min={0}
                        className="w-24"
                        disabled={isUnlimited}
                        value={
                          isUnlimited ? "" : String(formFeatures[key] ?? 0)
                        }
                        placeholder={isUnlimited ? "∞" : "0"}
                        onChange={(e) =>
                          setFormFeatures((prev) => ({
                            ...prev,
                            [key]: Math.max(
                              0,
                              parseInt(e.target.value, 10) || 0,
                            ),
                          }))
                        }
                      />
                    </div>
                  );
                })}
              </FieldGroup>
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
              onClick={confirmUpdate}
              disabled={submitting || !formName || !formBillingCycle}
            >
              {submitting ? tCommon("actions.saving") : t("editDialog.confirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
