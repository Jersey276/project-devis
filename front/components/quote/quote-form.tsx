"use client";

import { useMemo, useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import QuoteStepBasicInfo from "@/components/quote/steps/quote-step-basic-info";
import QuoteStepItems from "@/components/quote/steps/quote-step-items";
import QuoteStepSummary from "@/components/quote/steps/quote-step-summary";
import type { Quote, QuoteItem, QuoteVat } from "@/types/backend";

export type {
  Quote,
  QuoteItem,
  QuoteItems,
  QuoteStatus,
  QuoteVat,
} from "@/types/backend";

type QuoteFormProps = {
  quote?: Quote;
};

const STEP_LABELS = [
  "Informations de base",
  "Éléments du devis",
  "Résumé",
] as const;

const VAT_OPTIONS: QuoteVat[] = [
  { name: "TVA 0%", rate: 0 },
  { name: "TVA 5.5%", rate: 5.5 },
  { name: "TVA 10%", rate: 10 },
  { name: "TVA 20%", rate: 20 },
];

const VAT_OPTION_NAMES = VAT_OPTIONS.map((vat) => vat.name);

const EMPTY_CLIENTS: string[] = [];

function buildItem(): QuoteItem {
  return {
    id: crypto.randomUUID(),
    description: "",
    quantity: 1,
    unitPrice: 0,
    vat: VAT_OPTIONS[3],
  };
}

export default function QuoteForm({ quote }: QuoteFormProps) {
  const mode = quote ? "edit" : "create";
  const [step, setStep] = useState(0);
  const [projectName, setProjectName] = useState(quote?.name ?? "");
  const [clientId, setClientId] = useState(quote?.clientId ?? "");
  const [items, setItems] = useState<QuoteItem[]>(
    quote?.items?.length ? quote.items : [buildItem()],
  );

  const isReadonly =
    mode === "edit" && (quote?.status === "sent" || quote?.status === "signed");

  const totalAmount = useMemo(
    () =>
      items.reduce((acc, item) => {
        const lineBase = item.quantity * item.unitPrice;
        const lineTotal = lineBase + (lineBase * item.vat.rate) / 100;
        return acc + lineTotal;
      }, 0),
    [items],
  );

  const canGoNext =
    step === 0
      ? projectName.trim().length > 0
      : step === 1
        ? items.length > 0
        : true;

  const updateItem = <K extends keyof QuoteItem>(
    id: string,
    field: K,
    value: QuoteItem[K],
  ) => {
    setItems((previous) =>
      previous.map((item) =>
        item.id === id ? { ...item, [field]: value } : item,
      ),
    );
  };

  const removeItem = (id: string) => {
    setItems((previous) => {
      if (previous.length <= 1) {
        return previous;
      }
      return previous.filter((item) => item.id !== id);
    });
  };

  const submitLabel = mode === "create" ? "Créer le devis" : "Enregistrer";

  return (
    <Card>
      <CardHeader>
        <CardTitle>
          {mode === "create" ? "Création d’un devis" : `Devis ${quote?.name}`}
        </CardTitle>
        <CardDescription>
          {mode === "create"
            ? "Complète les étapes pour générer un nouveau devis."
            : isReadonly
              ? `Ce devis est en lecture seule (statut: ${quote?.status}).`
              : `Modification du devis (statut: ${quote?.status}).`}
        </CardDescription>
      </CardHeader>

      <CardContent className="space-y-6">
        <div className="grid gap-2 sm:grid-cols-3">
          {STEP_LABELS.map((label, index) => (
            <Button
              key={label}
              type="button"
              variant="ghost"
              className={`justify-start rounded-md border px-3 py-2 text-sm font-normal ${
                step === index ? "bg-muted" : ""
              }`}
              onClick={() => {
                if (mode === "edit") {
                  setStep(index);
                }
              }}
              disabled={mode !== "edit"}
            >
              Étape {index + 1} · {label}
            </Button>
          ))}
        </div>

        {step === 0 && (
          <QuoteStepBasicInfo
            projectName={projectName}
            clientId={clientId}
            isReadonly={isReadonly}
            emptyClients={EMPTY_CLIENTS}
            onProjectNameChange={setProjectName}
            onClientIdChange={setClientId}
          />
        )}

        {step === 1 && (
          <QuoteStepItems
            items={items}
            isReadonly={isReadonly}
            totalAmount={totalAmount}
            vatOptions={VAT_OPTIONS}
            onDescriptionChange={(id, value) =>
              updateItem(id, "description", value)
            }
            onQuantityChange={(id, value) => updateItem(id, "quantity", value)}
            onUnitPriceChange={(id, value) =>
              updateItem(id, "unitPrice", value)
            }
            onVatChange={(id, value) => updateItem(id, "vat", value)}
            onRemoveItem={removeItem}
            onAddItem={() => setItems((prev) => [...prev, buildItem()])}
          />
        )}

        {step === 2 && <QuoteStepSummary />}
      </CardContent>

      <CardFooter className="justify-between border-t">
        <Button
          type="button"
          variant="outline"
          onClick={() => setStep((current) => Math.max(0, current - 1))}
          disabled={step === 0}
        >
          Précédent
        </Button>

        <div className="flex gap-2">
          {step === 0 ? (
            <Button
              type="button"
              onClick={() =>
                setStep((current) =>
                  Math.min(STEP_LABELS.length - 1, current + 1),
                )
              }
              disabled={!canGoNext}
            >
              Suivant
            </Button>
          ) : (
            <Button type="button" disabled={isReadonly || items.length === 0}>
              {submitLabel}
            </Button>
          )}
        </div>
      </CardFooter>
    </Card>
  );
}
