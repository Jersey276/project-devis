"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { ChevronDownIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
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
import {
  continueQuote,
  dropQuote,
  negociateQuote,
  validateQuote,
} from "@/lib/services/quotes";
import type { ApiResult } from "@/lib/api";
import type { BackendQuoteState } from "@/types/backend";

// The set of transitions a quote can take from each state, mirroring the
// backend state machine (validate.go / negociate.go / drop.go / continue.go).
// Entering negotiation also sends the quote to the client by email.
// 'validated' is terminal → no entry.
type TransitionKey = "validate" | "negociate" | "drop" | "continue";

const TRANSITIONS: Record<BackendQuoteState, TransitionKey[]> = {
  draft: ["negociate", "drop"],
  negociation: ["validate", "drop"],
  validated: [],
  drop: ["continue"],
};

// Transitions that require an explicit confirmation popup.
const CONFIRM: ReadonlySet<TransitionKey> = new Set([
  "validate",
  "negociate",
  "drop",
]);

const ACTIONS: Record<TransitionKey, (id: string) => Promise<ApiResult>> = {
  validate: validateQuote,
  negociate: negociateQuote,
  drop: dropQuote,
  continue: continueQuote,
};

type Props = {
  quoteId: string;
  state: BackendQuoteState;
  disabled?: boolean;
  onChanged: (next: BackendQuoteState) => void;
  onError: (message: string) => void;
};

// Maps a transition to the resulting state, so the parent can update optimistically.
const RESULT_STATE: Record<TransitionKey, BackendQuoteState> = {
  validate: "validated",
  negociate: "negociation",
  drop: "drop",
  continue: "draft",
};

export default function QuoteStateDropdown({
  quoteId,
  state,
  disabled,
  onChanged,
  onError,
}: Props) {
  const t = useTranslations("quote.stateDropdown");
  const [pending, setPending] = useState<TransitionKey | null>(null);
  const [running, setRunning] = useState(false);

  const transitions = TRANSITIONS[state];
  if (transitions.length === 0) return null;

  async function run(key: TransitionKey) {
    setRunning(true);
    const { ok, body } = await ACTIONS[key](quoteId);
    setRunning(false);
    setPending(null);
    if (!ok || !body.success) {
      onError((body.message as string) ?? t("error"));
      return;
    }
    onChanged(RESULT_STATE[key]);
  }

  function onSelect(key: TransitionKey) {
    if (CONFIRM.has(key)) {
      setPending(key);
      return;
    }
    void run(key);
  }

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button type="button" variant="outline" disabled={disabled || running}>
            {t("trigger")}
            <ChevronDownIcon className="size-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          {transitions.map((key) => (
            <DropdownMenuItem
              key={key}
              variant={key === "drop" ? "destructive" : "default"}
              onSelect={() => onSelect(key)}
            >
              {t(`transitions.${key}`)}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>

      <AlertDialog
        open={pending !== null}
        onOpenChange={(open) => {
          if (!open) setPending(null);
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {pending ? t(`confirm.${pending}.title`) : ""}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {pending ? t(`confirm.${pending}.description`) : ""}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={running}>
              {t("confirm.cancel")}
            </AlertDialogCancel>
            <AlertDialogAction
              variant={pending === "drop" ? "destructive" : "default"}
              disabled={running}
              onClick={(e) => {
                e.preventDefault();
                if (pending) void run(pending);
              }}
            >
              {pending ? t(`confirm.${pending}.confirm`) : ""}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
