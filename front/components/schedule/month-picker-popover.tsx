"use client";

import { useMemo } from "react";
import { CalendarIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

type MonthPickerPopoverProps = {
  id?: string;
  name?: string;
  value: string;
  startYear: string;
  startMonthValue: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onStartYearChange: (year: string) => void;
  onStartMonthChange: (month: string) => void;
};

const MONTH_OPTIONS = [
  { value: "01", label: "Janvier" },
  { value: "02", label: "Fevrier" },
  { value: "03", label: "Mars" },
  { value: "04", label: "Avril" },
  { value: "05", label: "Mai" },
  { value: "06", label: "Juin" },
  { value: "07", label: "Juillet" },
  { value: "08", label: "Aout" },
  { value: "09", label: "Septembre" },
  { value: "10", label: "Octobre" },
  { value: "11", label: "Novembre" },
  { value: "12", label: "Decembre" },
];

export default function MonthPickerPopover({
  id = "schedule-start-month",
  name = "start_month",
  value,
  startYear,
  startMonthValue,
  open,
  onOpenChange,
  onStartYearChange,
  onStartMonthChange,
}: MonthPickerPopoverProps) {
  const startMonthLabel = useMemo(() => {
    if (!startYear || !startMonthValue) return "Choisir une annee et un mois";
    const monthLabel =
      MONTH_OPTIONS.find((option) => option.value === startMonthValue)?.label ??
      startMonthValue;
    return `${monthLabel} ${startYear}`;
  }, [startMonthValue, startYear]);

  const yearOptions = useMemo(() => {
    const currentYear = new Date().getFullYear();
    return Array.from({ length: 11 }, (_, index) =>
      String(currentYear - 3 + index),
    );
  }, []);

  function useCurrentMonth() {
    const now = new Date();
    onStartYearChange(String(now.getFullYear()));
    onStartMonthChange(String(now.getMonth() + 1).padStart(2, "0"));
  }

  return (
    <>
      <input type="hidden" name={name} value={value} readOnly />
      <Popover open={open} onOpenChange={onOpenChange}>
        <PopoverTrigger asChild>
          <Button
            id={id}
            type="button"
            variant="outline"
            className="w-full justify-between"
          >
            <span>{startMonthLabel}</span>
            <CalendarIcon className="size-4" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-80 space-y-3" align="start">
          <div className="grid grid-cols-2 gap-2">
            <div className="space-y-1">
              <p className="text-muted-foreground text-xs">Année</p>
              <Select value={startYear} onValueChange={onStartYearChange}>
                <SelectTrigger
                  data-slot="schedule-start-year-trigger"
                  className="w-full"
                >
                  <SelectValue placeholder="Choisir" />
                </SelectTrigger>
                <SelectContent>
                  {yearOptions.map((year) => (
                    <SelectItem key={year} value={year}>
                      {year}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1">
              <p className="text-muted-foreground text-xs">Mois</p>
              <Select
                value={startMonthValue}
                onValueChange={onStartMonthChange}
              >
                <SelectTrigger
                  data-slot="schedule-start-month-trigger"
                  className="w-full"
                >
                  <SelectValue placeholder="Choisir" />
                </SelectTrigger>
                <SelectContent>
                  {MONTH_OPTIONS.map((month) => (
                    <SelectItem key={month.value} value={month.value}>
                      {month.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="flex justify-between gap-2">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={useCurrentMonth}
            >
              Mois courant
            </Button>
            <Button
              type="button"
              size="sm"
              onClick={() => onOpenChange(false)}
              disabled={!value}
            >
              Valider
            </Button>
          </div>
        </PopoverContent>
      </Popover>
    </>
  );
}
