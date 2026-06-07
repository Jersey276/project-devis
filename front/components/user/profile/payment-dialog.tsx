"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { loadStripe } from "@stripe/stripe-js";
import {
  Elements,
  PaymentElement,
  useStripe,
  useElements,
} from "@stripe/react-stripe-js";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { toast } from "sonner";
import { createPaymentIntent } from "@/lib/services/subscriptions";
import type { BackendPlan } from "@/types/backend";

const stripePromise = process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY
  ? loadStripe(process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY)
  : null;

type BillingDetails = {
  email?: string;
  phone?: string;
  name?: string;
};

type PaymentDialogProps = {
  open: boolean;
  onOpenChange: (v: boolean) => void;
  plan: BackendPlan | null;
  onSuccess: () => void;
  billingDetails?: BillingDetails;
};

function CheckoutForm({
  onSuccess,
  onClose,
  billingDetails,
}: {
  onSuccess: () => void;
  onClose: () => void;
  billingDetails?: BillingDetails;
}) {
  const t = useTranslations("subscription.paymentDialog");
  const tCommon = useTranslations("common");
  const stripe = useStripe();
  const elements = useElements();
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!stripe || !elements) return;

    setSubmitting(true);
    const { error } = await stripe.confirmPayment({
      elements,
      confirmParams: {
        return_url: window.location.href,
      },
      redirect: "if_required",
    });

    if (error) {
      toast.error(error.message ?? t("error"));
      setSubmitting(false);
      return;
    }

    toast.success(t("success"));
    // Give the webhook ~2s to be processed before refreshing
    setTimeout(() => {
      onSuccess();
      onClose();
      setSubmitting(false);
    }, 2000);
  }

  return (
    <form onSubmit={handleSubmit} className="grid gap-4">
      <PaymentElement
        options={{
          defaultValues: {
            billingDetails: {
              name: billingDetails?.name || undefined,
              email: billingDetails?.email || undefined,
              phone: billingDetails?.phone || undefined,
            },
          },
        }}
      />
      <DialogFooter>
        <Button type="button" variant="outline" onClick={onClose} disabled={submitting}>
          {tCommon("actions.cancel")}
        </Button>
        <Button type="submit" disabled={!stripe || submitting}>
          {submitting ? tCommon("actions.saving") : t("confirmButton")}
        </Button>
      </DialogFooter>
    </form>
  );
}

export default function PaymentDialog({
  open,
  onOpenChange,
  plan,
  onSuccess,
  billingDetails,
}: PaymentDialogProps) {
  const t = useTranslations("subscription.paymentDialog");
  const [clientSecret, setClientSecret] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const loading = open && !clientSecret && !error;

  useEffect(() => {
    if (!open || !plan || clientSecret || error) return;

    createPaymentIntent(plan.plan_id).then(({ ok, body }) => {
      if (ok && body.client_secret) {
        setClientSecret(body.client_secret as string);
      } else {
        setError(body.message ?? t("processingError"));
      }
    });
  }, [open, plan, clientSecret, error, t]);

  function handleOpenChange(v: boolean) {
    if (!v) {
      setClientSecret(null);
      setError(null);
    }
    onOpenChange(v);
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>
            {plan ? t("title", { name: plan.name }) : t("title", { name: "" })}
          </DialogTitle>
        </DialogHeader>

        {loading && (
          <p className="text-sm text-muted-foreground">{t("loading")}</p>
        )}

        {error && (
          <p className="text-sm text-destructive">{error}</p>
        )}

        {clientSecret && (
          <Elements
            stripe={stripePromise}
            options={{ clientSecret, locale: "fr" }}
          >
            <CheckoutForm
              onSuccess={onSuccess}
              onClose={() => handleOpenChange(false)}
              billingDetails={billingDetails}
            />
          </Elements>
        )}
      </DialogContent>
    </Dialog>
  );
}
