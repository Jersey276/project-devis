"use client";

import { useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Loader2Icon } from "lucide-react";
import { listFees } from "@/lib/services/fees";
import type { BackendFee } from "@/types/backend";

type SelectFeePopoverProps = {
  children: React.ReactNode;
  disabled?: boolean;
  onSelect: (fee: BackendFee) => Promise<void> | void;
};

// SelectFeePopover lists the user's fee catalog and yields the chosen fee so the
// caller can snapshot it onto a quote line or subline. Mirrors
// SelectLineTemplatePopover.
export default function SelectFeePopover({
  children,
  disabled,
  onSelect,
}: SelectFeePopoverProps) {
  const t = useTranslations("fees.selectPopover");
  const tCategories = useTranslations("fees.categories");
  const [open, setOpen] = useState(false);
  const [fees, setFees] = useState<BackendFee[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectingId, setSelectingId] = useState<string | null>(null);
  const [search, setSearch] = useState("");

  function handleSetOpen(newOpen: boolean) {
    if (newOpen) setSearch("");
    if (!newOpen) setLoading(true);
    setOpen(newOpen);
  }

  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    listFees().then(({ ok, body }) => {
      if (cancelled) return;
      setLoading(false);
      if (ok && Array.isArray(body.fees)) {
        setFees(body.fees as BackendFee[]);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [open]);

  const filtered = useMemo(
    () =>
      search.trim() === ""
        ? fees
        : fees.filter((f) =>
            f.name.toLowerCase().includes(search.toLowerCase()),
          ),
    [fees, search],
  );

  async function handleSelect(fee: BackendFee) {
    setSelectingId(fee.fee_id);
    try {
      await onSelect(fee);
      handleSetOpen(false);
    } finally {
      setSelectingId(null);
    }
  }

  return (
    <Popover open={open} onOpenChange={disabled ? undefined : handleSetOpen}>
      <PopoverTrigger asChild>{children}</PopoverTrigger>
      <PopoverContent className="w-72 p-2">
        {loading ? (
          <div className="flex justify-center py-4">
            <Loader2Icon className="text-muted-foreground size-5 animate-spin" />
          </div>
        ) : fees.length === 0 ? (
          <p className="text-muted-foreground py-2 text-center text-sm">
            {t("empty")}
          </p>
        ) : (
          <div className="flex flex-col gap-1">
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={t("searchPlaceholder")}
              className="mb-1 h-8 text-sm"
            />
            {filtered.length === 0 ? (
              <p className="text-muted-foreground py-2 text-center text-sm">
                {t("noResults")}
              </p>
            ) : (
              <div className="max-h-[calc(15*2.25rem)] overflow-y-auto">
                {filtered.map((fee) => (
                  <Button
                    key={fee.fee_id}
                    variant="ghost"
                    className="h-auto w-full flex-col items-start gap-0.5 py-1.5"
                    disabled={selectingId !== null}
                    onClick={() => handleSelect(fee)}
                  >
                    <span className="flex w-full items-center gap-2">
                      {selectingId === fee.fee_id && (
                        <Loader2Icon className="size-4 animate-spin" />
                      )}
                      <span className="truncate font-medium">{fee.name}</span>
                    </span>
                    <span className="text-muted-foreground text-xs">
                      {tCategories(fee.category)} ·{" "}
                      {(fee.unit_price / 100).toFixed(2)} €
                      {fee.unit ? ` / ${fee.unit}` : ""}
                    </span>
                  </Button>
                ))}
              </div>
            )}
          </div>
        )}
      </PopoverContent>
    </Popover>
  );
}
