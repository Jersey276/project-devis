"use client";

import * as React from "react";
import { useTranslations } from "next-intl";
import { CalendarIcon } from "lucide-react";
import type { DateRange } from "react-day-picker";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

export type DateRangePickerProps = {
  from: string;
  to: string;
  onValueChange: (from: string, to: string) => void;
};

function isoToDate(s: string): Date | undefined {
  if (!s) return undefined;
  const d = new Date(s + "T00:00:00Z");
  return isNaN(d.getTime()) ? undefined : d;
}

function dateToISO(d: Date): string {
  return d.toISOString().split("T")[0];
}

function formatDateFR(s: string): string {
  const d = isoToDate(s);
  if (!d) return "";
  return d.toLocaleDateString("fr-FR", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    timeZone: "UTC",
  });
}

export function DateRangePicker({ from, to, onValueChange }: DateRangePickerProps) {
  const t = useTranslations("common.dateRangePicker");
  const [open, setOpen] = React.useState(false);
  const [range, setRange] = React.useState<DateRange | undefined>({
    from: isoToDate(from),
    to: isoToDate(to),
  });

  React.useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- sync with parent-controlled props (filter reset)
    setRange({ from: isoToDate(from), to: isoToDate(to) });
  }, [from, to]);

  function handleSelect(r: DateRange | undefined) {
    setRange(r);
    onValueChange(
      r?.from ? dateToISO(r.from) : "",
      r?.to ? dateToISO(r.to) : "",
    );
    if (r?.from && r?.to) setOpen(false);
  }

  function buildLabel(): string {
    if (from && to) return `${formatDateFR(from)} – ${formatDateFR(to)}`;
    if (from) return `${t("from")} ${formatDateFR(from)}`;
    if (to) return `${t("to")} ${formatDateFR(to)}`;
    return t("placeholder");
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className="w-full justify-start font-normal"
          data-empty={!from && !to}
        >
          <CalendarIcon className="size-4 shrink-0" />
          <span className="truncate text-xs">{buildLabel()}</span>
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <Calendar
          mode="range"
          selected={range}
          onSelect={handleSelect}
          numberOfMonths={1}
        />
        {(from || to) && (
          <div className="border-t px-3 py-2">
            <Button
              variant="ghost"
              size="sm"
              className="w-full text-xs"
              onClick={() => {
                setRange(undefined);
                onValueChange("", "");
              }}
            >
              {t("clear")}
            </Button>
          </div>
        )}
      </PopoverContent>
    </Popover>
  );
}
