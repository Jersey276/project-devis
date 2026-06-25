"use client";

import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { CoinsIcon, PlusIcon, Trash2Icon } from "lucide-react";
import type { BackendFee, QuoteLineData } from "@/types/backend";
import SelectFeePopover from "@/components/fees/select-fee-popover";
import type { QuoteItemRow } from "./quote-step-items";

type Props = {
  item: QuoteItemRow;
  isReadonly: boolean;
  onSublineAdd?: (lineId: string) => void;
  onSublineChange?: (
    lineId: string,
    index: number,
    patch: Partial<NonNullable<QuoteLineData["sublines"]>[number]>,
  ) => void;
  onSublineRemove?: (lineId: string, index: number) => void;
  onAddFeeSubline?: (lineId: string, fee: BackendFee) => void;
};

export default function LineDetailedSublines({
  item,
  isReadonly,
  onSublineAdd,
  onSublineChange,
  onSublineRemove,
  onAddFeeSubline,
}: Props) {
  const t = useTranslations("quote.steps.items");
  const kind = item.data.kind ?? "line";
  const sublines = item.data.sublines ?? [];

  if (kind !== "detailed") return null;
  if (isReadonly && sublines.length === 0) return null;

  return (
    <TableRow key={`${item.lineId}-sublines`}>
      <TableCell colSpan={7} className="p-0 pb-2 pl-8">
        <div className="overflow-hidden rounded-md border">
          {sublines.length > 0 && (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("sublines.name")}</TableHead>
                  <TableHead>{t("sublines.quantity")}</TableHead>
                  <TableHead>{t("sublines.unit")}</TableHead>
                  <TableHead>{t("sublines.unitPrice")}</TableHead>
                  <TableHead>{t("sublines.option")}</TableHead>
                  <TableHead className="w-10" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {sublines.map((subline, idx) => {
                  // A subline added from the catalog carries a fee_id; its
                  // name/unit/price are catalog-driven, so lock them like a
                  // fee line. Quantity and option stay editable.
                  const sublineIsFee = !!subline.fee_id;
                  return (
                    <TableRow key={subline._key ?? String(idx)}>
                      <TableCell>
                        <Input
                          value={subline.name}
                          onChange={(e) =>
                            onSublineChange?.(item.lineId, idx, {
                              name: e.target.value,
                            })
                          }
                          disabled={isReadonly || sublineIsFee}
                          placeholder={t("lineNamePlaceholder")}
                        />
                      </TableCell>
                      <TableCell>
                        <Input
                          type="number"
                          min={0}
                          value={subline.quantity}
                          onChange={(e) =>
                            onSublineChange?.(item.lineId, idx, {
                              quantity: e.target.value,
                            })
                          }
                          disabled={isReadonly}
                        />
                      </TableCell>
                      <TableCell>
                        <Input
                          value={subline.unit ?? ""}
                          onChange={(e) =>
                            onSublineChange?.(item.lineId, idx, {
                              unit: e.target.value || undefined,
                            })
                          }
                          disabled={isReadonly || sublineIsFee}
                          placeholder={t("sublines.unitPlaceholder")}
                        />
                      </TableCell>
                      <TableCell>
                        <Input
                          type="number"
                          min={0}
                          step="0.01"
                          value={subline.unit_price / 100}
                          onChange={(e) =>
                            onSublineChange?.(item.lineId, idx, {
                              unit_price: Math.round(
                                Number(e.target.value) * 100,
                              ),
                            })
                          }
                          disabled={isReadonly || sublineIsFee}
                        />
                      </TableCell>
                      <TableCell>
                        <Checkbox
                          checked={!!subline.option}
                          onCheckedChange={(checked) =>
                            onSublineChange?.(item.lineId, idx, {
                              option: checked === true,
                            })
                          }
                          disabled={isReadonly}
                        />
                      </TableCell>
                      <TableCell>
                        {!isReadonly && (
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            aria-label={t("sublines.deleteAria")}
                            onClick={() => onSublineRemove?.(item.lineId, idx)}
                          >
                            <Trash2Icon className="size-4" />
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
          {!isReadonly && (onSublineAdd || onAddFeeSubline) && (
            <div className="flex">
              {onSublineAdd && (
                <Button
                  type="button"
                  variant="ghost"
                  className="h-auto flex-1 p-0"
                  onClick={() => onSublineAdd(item.lineId)}
                  aria-label={t("sublines.addAria")}
                >
                  <Skeleton className="flex h-8 w-full items-center justify-center">
                    <PlusIcon className="size-4" />
                  </Skeleton>
                </Button>
              )}
              {onAddFeeSubline && (
                <SelectFeePopover
                  onSelect={(fee) => onAddFeeSubline(item.lineId, fee)}
                >
                  <Button
                    type="button"
                    variant="ghost"
                    className="h-auto p-0"
                    aria-label={t("sublines.addFeeAria")}
                  >
                    <Skeleton className="flex h-8 w-12 items-center justify-center">
                      <CoinsIcon className="size-4" />
                    </Skeleton>
                  </Button>
                </SelectFeePopover>
              )}
            </div>
          )}
        </div>
      </TableCell>
    </TableRow>
  );
}
