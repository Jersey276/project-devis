"use client";

import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { CoinsIcon, LayoutTemplateIcon, Loader2Icon, PlusIcon } from "lucide-react";
import type { BackendFee, QuoteLineData } from "@/types/backend";
import SelectFeePopover from "@/components/fees/select-fee-popover";
import SelectLineTemplatePopover from "@/components/template/select-line-template-popover";

const CREATABLE_KINDS: Array<QuoteLineData["kind"]> = [
  "line",
  "text",
  "group",
  "detailed",
];

type Props = {
  isAdding: boolean;
  kindLabel: (kind: QuoteLineData["kind"]) => string;
  onAddItem: (kind?: QuoteLineData["kind"]) => void;
  onAddFeeItem?: (fee: BackendFee) => void | Promise<void>;
  onAddItemFromTemplate?: (templateId: string) => Promise<void>;
};

export default function AddItemControls({
  isAdding,
  kindLabel,
  onAddItem,
  onAddFeeItem,
  onAddItemFromTemplate,
}: Props) {
  const t = useTranslations("quote.steps.items");

  return (
    <div className="flex gap-2">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            type="button"
            variant="ghost"
            className="h-auto flex-1 p-0"
            disabled={isAdding}
            aria-label={t("addAria")}
            title={t("addTypeAria")}
          >
            <Skeleton className="flex h-14 w-full items-center justify-center">
              {isAdding ? (
                <Loader2Icon className="size-7 animate-spin" />
              ) : (
                <PlusIcon className="size-7" />
              )}
            </Skeleton>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start">
          {CREATABLE_KINDS.map((kind) => (
            <DropdownMenuItem key={kind} onClick={() => onAddItem(kind)}>
              {kindLabel(kind)}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
      {onAddFeeItem && (
        <SelectFeePopover
          disabled={isAdding}
          onSelect={(fee) => onAddFeeItem(fee)}
        >
          <Button
            type="button"
            variant="ghost"
            className="h-auto p-0"
            disabled={isAdding}
            aria-label={t("addFeeAria")}
          >
            <Skeleton className="flex h-14 w-14 items-center justify-center">
              <CoinsIcon className="size-5" />
            </Skeleton>
          </Button>
        </SelectFeePopover>
      )}
      {onAddItemFromTemplate && (
        <SelectLineTemplatePopover
          disabled={isAdding}
          onSelect={onAddItemFromTemplate}
        >
          <Button
            type="button"
            variant="ghost"
            className="h-auto p-0"
            disabled={isAdding}
            aria-label={t("addFromTemplateAria")}
          >
            <Skeleton className="flex h-14 w-14 items-center justify-center">
              <LayoutTemplateIcon className="size-5" />
            </Skeleton>
          </Button>
        </SelectLineTemplatePopover>
      )}
    </div>
  );
}
