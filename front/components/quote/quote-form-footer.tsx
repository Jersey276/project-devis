"use client";

import { useTranslations } from "next-intl";
import { CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

type Props = {
  step: number;
  stepCount: number;
  creating: boolean;
  canGoNextFromStep1: boolean;
  onPrev: () => void;
  onNextFromStep1: () => void;
  onNextStep: () => void;
  onFinish: () => void;
};

export default function QuoteFormFooter({
  step,
  stepCount,
  creating,
  canGoNextFromStep1,
  onPrev,
  onNextFromStep1,
  onNextStep,
  onFinish,
}: Props) {
  const t = useTranslations("quote.form");

  return (
    <CardFooter className="justify-between border-t">
      <Button
        type="button"
        variant="outline"
        onClick={onPrev}
        disabled={step === 0}
      >
        {t("prev")}
      </Button>

      <div className="flex gap-2">
        {step === 0 ? (
          <Button
            type="button"
            onClick={onNextFromStep1}
            disabled={!canGoNextFromStep1}
          >
            {creating ? t("creating") : t("next")}
          </Button>
        ) : step < stepCount - 1 ? (
          <Button type="button" onClick={onNextStep}>
            {t("next")}
          </Button>
        ) : (
          <Button type="button" variant="outline" onClick={onFinish}>
            {t("finish")}
          </Button>
        )}
      </div>
    </CardFooter>
  );
}
