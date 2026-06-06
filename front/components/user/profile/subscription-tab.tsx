"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { toast } from "sonner";
import {
  getMySubscription,
  listPlans,
  cancelSubscription,
} from "@/lib/services/subscriptions";
import PaymentDialog from "./payment-dialog";
import type { BackendSubscription, BackendPlan } from "@/types/backend";

type SubscriptionTabProps = {
  userId: string;
  readOnly?: boolean;
  email?: string;
  phone?: string;
  name?: string;
};

function formatDate(value: string | null | undefined): string {
  if (!value) return "";
  const d = new Date(value);
  if (isNaN(d.getTime())) return "";
  return new Intl.DateTimeFormat("fr-FR", { dateStyle: "long" }).format(d);
}

function formatPrice(priceCents: number, billingCycle: string): string {
  if (priceCents === 0) return "Gratuit";
  const euros = (priceCents / 100).toLocaleString("fr-FR", {
    style: "currency",
    currency: "EUR",
  });
  return billingCycle === "yearly" ? `${euros}/an` : `${euros}/mois`;
}

export default function SubscriptionTab({ readOnly, email, phone, name }: SubscriptionTabProps) {
  const t = useTranslations("subscription");
  const tCommon = useTranslations("common");

  const [subscription, setSubscription] = useState<BackendSubscription | null>(null);
  const [plans, setPlans] = useState<BackendPlan[]>([]);
  const [loadedFor, setLoadedFor] = useState<number | null>(null);
  const [reloadKey, setReloadKey] = useState(0);
  const loading = loadedFor !== reloadKey;
  const [selectedPlan, setSelectedPlan] = useState<BackendPlan | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [cancelling, setCancelling] = useState(false);

  const reload = () => setReloadKey((k) => k + 1);

  useEffect(() => {
    let cancelled = false;

    Promise.all([getMySubscription(), listPlans()]).then(
      ([subResult, plansResult]) => {
        if (cancelled) return;
        if (subResult.ok && subResult.body.subscription) {
          setSubscription(subResult.body.subscription as BackendSubscription);
        }
        if (plansResult.ok && Array.isArray(plansResult.body.plans)) {
          setPlans(plansResult.body.plans as BackendPlan[]);
        }
        setLoadedFor(reloadKey);
      },
    );

    return () => {
      cancelled = true;
    };
  }, [reloadKey]);

  async function handleCancel() {
    setCancelling(true);
    const { ok, body } = await cancelSubscription();
    setCancelling(false);
    if (ok && body.success) {
      toast.success(t("tab.cancelSuccess"));
      reload();
    } else {
      toast.error(body.message ?? t("tab.cancelError"));
    }
  }

  function openPaymentDialog(plan: BackendPlan) {
    setSelectedPlan(plan);
    setDialogOpen(true);
  }

  const currentTier = subscription?.tier ?? "free";

  const currentPlanInfo = () => {
    if (!subscription || currentTier === "free") {
      return <p className="text-sm text-muted-foreground">{t("tab.freePlan")}</p>;
    }
    if (subscription.cancel_at_period_end) {
      return (
        <p className="text-sm text-amber-600">
          {t("tab.cancelPending", { date: formatDate(subscription.current_period_end) })}
        </p>
      );
    }
    return (
      <p className="text-sm text-muted-foreground">
        {t("tab.renewsOn", { date: formatDate(subscription.current_period_end) })}
      </p>
    );
  };

  if (loading) {
    return <p className="text-sm text-muted-foreground">{tCommon("actions.loading")}</p>;
  }

  return (
    <div className="grid gap-6">
      <div className="flex items-center gap-3">
        <span className="text-sm font-medium">{t("tab.currentPlan")}</span>
        <Badge variant={currentTier === "enterprise" ? "default" : currentTier === "pro" ? "secondary" : "outline"}>
          {currentTier.charAt(0).toUpperCase() + currentTier.slice(1)}
        </Badge>
        {currentPlanInfo()}
      </div>

      {!subscription?.cancel_at_period_end && currentTier !== "free" && !readOnly && (
        <Button
          variant="outline"
          size="sm"
          className="w-fit text-destructive border-destructive/30 hover:bg-destructive/5"
          onClick={handleCancel}
          disabled={cancelling}
        >
          {cancelling ? tCommon("actions.saving") : t("tab.cancelCta")}
        </Button>
      )}

      <div className="grid gap-4 sm:grid-cols-3">
        {plans.map((plan) => {
          const isCurrent = plan.tier === currentTier;
          const isUpgrade =
            (currentTier === "free" && plan.tier !== "free") ||
            (currentTier === "pro" && plan.tier === "enterprise");

          return (
            <Card
              key={plan.plan_id}
              className={isCurrent ? "border-primary ring-1 ring-primary" : ""}
            >
              <CardHeader className="pb-2">
                <CardTitle className="text-base">{plan.name}</CardTitle>
                <p className="text-sm font-semibold text-foreground">
                  {formatPrice(plan.price_cents, plan.billing_cycle)}
                </p>
              </CardHeader>
              <CardContent>
                {isCurrent && (
                  <Badge variant="outline" className="text-xs">Plan actuel</Badge>
                )}
                {isUpgrade && !readOnly && (
                  <Button
                    size="sm"
                    className="mt-2 w-full"
                    onClick={() => openPaymentDialog(plan)}
                  >
                    {t("tab.subscribeCta")}
                  </Button>
                )}
              </CardContent>
            </Card>
          );
        })}
      </div>

      <PaymentDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        plan={selectedPlan}
        onSuccess={reload}
        billingDetails={{ email, phone, name }}
      />
    </div>
  );
}
