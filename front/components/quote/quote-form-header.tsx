"use client";

import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { BookmarkIcon, CalendarIcon, CheckIcon, DownloadIcon, XIcon } from "lucide-react";
import GenerateInvoiceFromQuoteButton from "@/components/invoice/generate-invoice-from-quote-button";
import QuoteStateDropdown from "@/components/quote/quote-state-dropdown";
import type { BackendQuoteState } from "@/types/backend";

const STATE_BADGE_VARIANT: Record<
  BackendQuoteState,
  "default" | "secondary" | "destructive"
> = {
  draft: "secondary",
  negociation: "default",
  validated: "default",
  drop: "destructive",
  accepted: "default",
  refused: "destructive",
};

type Props = {
  quoteId?: string;
  projectName: string;
  quoteState: BackendQuoteState;
  isCreate: boolean;
  isReadonly: boolean;
  isCustomer: boolean;
  isExporting: boolean;
  onExport: () => void;
  onSaveTemplate: () => void;
  onCreateSchedule: () => void;
  onStateChanged: (next: BackendQuoteState) => void;
  onAccept?: () => void;
  onRefuse?: () => void;
};

export default function QuoteFormHeader({
  quoteId,
  projectName,
  quoteState,
  isCreate,
  isReadonly,
  isCustomer,
  isExporting,
  onExport,
  onSaveTemplate,
  onCreateSchedule,
  onStateChanged,
  onAccept,
  onRefuse,
}: Props) {
  const t = useTranslations("quote.form");
  const tStatus = useTranslations("status.quote");

  return (
    <CardHeader className="flex flex-row items-start justify-between gap-4">
      <div className="space-y-1.5">
        <CardTitle className="flex items-center gap-2">
          {isCreate
            ? t("createTitle")
            : t("editTitlePlaceholder", { name: projectName || "…" })}
          {!isCreate && (
            <Badge
              data-slot="quote-state-badge"
              variant={STATE_BADGE_VARIANT[quoteState]}
            >
              {tStatus(quoteState)}
            </Badge>
          )}
        </CardTitle>
        <CardDescription>
          {isCreate ? t("createDescription") : t("editDescription")}
        </CardDescription>
      </div>
      {!isCreate && (
        <Button
          type="button"
          variant="outline"
          disabled={isExporting}
          onClick={onExport}
        >
          <DownloadIcon className="size-4" />
          {t("exportButton")}
        </Button>
      )}
      {!isCreate && quoteId ? (
        <GenerateInvoiceFromQuoteButton
          quoteId={quoteId}
          validated={quoteState === "validated"}
          onError={(message) => toast.error(message)}
        />
      ) : null}
      {!isCreate && !isReadonly && (
        <Button type="button" variant="outline" onClick={onSaveTemplate}>
          <BookmarkIcon className="size-4" />
          {t("saveAsTemplateButton")}
        </Button>
      )}
      {!isCreate && !isCustomer && (
        <Button type="button" variant="outline" onClick={onCreateSchedule}>
          <CalendarIcon className="size-4" />
          Créer un échéancier
        </Button>
      )}
      {!isCustomer && !isCreate && quoteId && (
        <QuoteStateDropdown
          quoteId={quoteId}
          state={quoteState}
          onChanged={onStateChanged}
          onError={(message) => toast.error(message)}
        />
      )}
      {isCustomer && !isCreate && quoteState === "negociation" && (
        <>
          <Button type="button" variant="outline" onClick={onRefuse}>
            <XIcon className="size-4" />
            {tStatus("refused")}
          </Button>
          <Button type="button" onClick={onAccept}>
            <CheckIcon className="size-4" />
            {tStatus("accepted")}
          </Button>
        </>
      )}
    </CardHeader>
  );
}
