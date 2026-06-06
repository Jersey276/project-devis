"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
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

function centsToEuros(cents: number): string {
  return (cents / 100).toFixed(2);
}

function eurosToCents(euros: string): number {
  return Math.round(parseFloat(euros || "0") * 100);
}

function featuresAsString(features: BackendPlan["features"]): string {
  if (typeof features === "string") return features;
  try {
    return JSON.stringify(features, null, 2);
  } catch {
    return "{}";
  }
}

export default function PlansTable() {
  const t = useTranslations("admin.plans");
  const tCommon = useTranslations("common");

  const [plans, setPlans] = useState<BackendPlan[]>([]);
  const [editing, setEditing] = useState<BackendPlan | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [reloadKey, setReloadKey] = useState(0);

  const [formName, setFormName] = useState("");
  const [formPriceEuros, setFormPriceEuros] = useState("");
  const [formBillingCycle, setFormBillingCycle] = useState<string>("none");
  const [formStripePriceId, setFormStripePriceId] = useState("");
  const [formFeatures, setFormFeatures] = useState("");

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
    setFormFeatures(featuresAsString(plan.features));
    setDialogOpen(true);
  }

  async function confirmUpdate() {
    if (!editing) return;

    let parsedFeatures: string;
    try {
      JSON.parse(formFeatures || "{}");
      parsedFeatures = formFeatures || "{}";
    } catch {
      toast.error(t("editDialog.invalidJson"));
      return;
    }

    setSubmitting(true);
    try {
      const { ok, body } = await updatePlan(editing.plan_id, {
        name: formName,
        price_cents: eurosToCents(formPriceEuros),
        billing_cycle: formBillingCycle,
        stripe_price_id: formStripePriceId,
        features: parsedFeatures,
      });
      if (ok && body.success) {
        toast.success(t("editDialog.successToast"));
        setReloadKey((k) => k + 1);
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
              <FieldLabel htmlFor="plan_name">{t("editDialog.nameLabel")}</FieldLabel>
              <Input
                id="plan_name"
                value={formName}
                onChange={(e) => setFormName(e.target.value)}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="plan_price">{t("editDialog.priceLabelEuros")}</FieldLabel>
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
              <Select value={formBillingCycle} onValueChange={setFormBillingCycle}>
                <SelectTrigger id="plan_billing_cycle" className="w-full">
                  <SelectValue placeholder={t("editDialog.billingCyclePlaceholder")} />
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
              <FieldDescription>{t("editDialog.stripePriceIdHint")}</FieldDescription>
            </Field>
            <Field>
              <FieldLabel htmlFor="plan_features">
                {t("editDialog.featuresLabel")}
              </FieldLabel>
              <Textarea
                id="plan_features"
                rows={4}
                className="font-mono text-xs"
                value={formFeatures}
                onChange={(e) => setFormFeatures(e.target.value)}
              />
              <FieldDescription>{t("editDialog.featuresHint")}</FieldDescription>
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
