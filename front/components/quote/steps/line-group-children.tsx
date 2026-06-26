"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Table, TableBody, TableCell, TableRow } from "@/components/ui/table";
import { Loader2Icon, PlusIcon } from "lucide-react";
import type { QuoteLineData } from "@/types/backend";
import type { QuoteItemRow } from "./quote-step-items";

const CREATABLE_KINDS: Array<QuoteLineData["kind"]> = [
  "line",
  "text",
  "group",
  "detailed",
];

type Props = {
  item: QuoteItemRow;
  isReadonly: boolean;
  isAdding: boolean;
  renderedChildren: React.ReactNode;
  kindLabel: (kind: QuoteLineData["kind"]) => string;
  onAddChildItem?: (parentLineId: string, kind?: QuoteLineData["kind"]) => void;
};

export default function LineGroupChildren({
  item,
  isReadonly,
  isAdding,
  renderedChildren,
  kindLabel,
  onAddChildItem,
}: Props) {
  const t = useTranslations("quote.steps.items");
  const kind = item.data.kind ?? "line";

  if (kind !== "group") return null;

  const hasChildren = React.Children.count(renderedChildren) > 0;
  if (isReadonly && !hasChildren) return null;

  return (
    <TableRow key={`${item.lineId}-group-children`}>
      <TableCell colSpan={7} className="p-0 pb-2 pl-8">
        <div className="overflow-hidden rounded-md border">
          {hasChildren && (
            <Table>
              <TableBody>{renderedChildren}</TableBody>
            </Table>
          )}
          {!isReadonly && onAddChildItem && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  type="button"
                  variant="ghost"
                  className="h-auto w-full p-0"
                  disabled={isAdding}
                  aria-label={t("addLineInGroupAria")}
                  title={t("addChildTypeAria")}
                >
                  <Skeleton className="flex h-8 w-full items-center justify-center">
                    {isAdding ? (
                      <Loader2Icon className="size-4 animate-spin" />
                    ) : (
                      <PlusIcon className="size-4" />
                    )}
                  </Skeleton>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start">
                {CREATABLE_KINDS.map((k) => (
                  <DropdownMenuItem
                    key={k}
                    onClick={() => onAddChildItem(item.lineId, k)}
                  >
                    {kindLabel(k)}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
      </TableCell>
    </TableRow>
  );
}
