"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Input } from "@/components/ui/input";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Button } from "@/components/ui/button";
import { SelectCombobox } from "@/components/ui/select-combobox";
import { FilterSidebar, FilterSidebarSection } from "@/components/ui/filter-sidebar";
import { CalendarIcon } from "lucide-react";
import type { DateRange } from "react-day-picker";

export type LogFilters = {
  search: string;
  resp_statuses: number[];
  date_from: string;
  date_to: string;
};

export const emptyLogFilters: LogFilters = {
  search: "",
  resp_statuses: [],
  date_from: "",
  date_to: "",
};

const HTTP_STATUS_ITEMS = [
  "200", "201", "204", "301", "302",
  "400", "401", "403", "404", "409", "422", "429",
  "500", "502", "503",
].map((s) => ({ value: s, label: s }));

function dateToISO(d: Date): string {
  return d.toISOString().split("T")[0];
}

function isoToDate(s: string): Date | undefined {
  if (!s) return undefined;
  const d = new Date(s + "T00:00:00Z");
  return isNaN(d.getTime()) ? undefined : d;
}

function formatDateFR(s: string): string {
  const d = isoToDate(s);
  if (!d) return "";
  return d.toLocaleDateString("fr-FR", { day: "2-digit", month: "2-digit", year: "numeric", timeZone: "UTC" });
}

type LogsFiltersProps = {
  filters: LogFilters;
  onChange: (filters: LogFilters) => void;
};

function activeCount(f: LogFilters): number {
  return (f.resp_statuses.length > 0 ? 1 : 0) + (f.date_from || f.date_to ? 1 : 0);
}

function buildDateLabel(f: LogFilters, t: ReturnType<typeof useTranslations>): string {
  if (f.date_from && f.date_to) return `${formatDateFR(f.date_from)} – ${formatDateFR(f.date_to)}`;
  if (f.date_from) return `${t("from")} ${formatDateFR(f.date_from)}`;
  if (f.date_to) return `${t("to")} ${formatDateFR(f.date_to)}`;
  return t("datePlaceholder");
}

export default function LogsFilters({ filters, onChange }: LogsFiltersProps) {
  const t = useTranslations("admin.logs.filters");
  const tCommon = useTranslations("common.filterSidebar");
  const [dateRange, setDateRange] = useState<DateRange | undefined>({
    from: isoToDate(filters.date_from),
    to: isoToDate(filters.date_to),
  });
  const [dateOpen, setDateOpen] = useState(false);

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...filters, search: e.target.value });
  };

  const handleDateSelect = (range: DateRange | undefined) => {
    setDateRange(range);
    onChange({
      ...filters,
      date_from: range?.from ? dateToISO(range.from) : "",
      date_to: range?.to ? dateToISO(range.to) : "",
    });
    if (range?.from && range?.to) {
      setDateOpen(false);
    }
  };

  const handleReset = () => {
    setDateRange(undefined);
    onChange(emptyLogFilters);
  };

  return (
    <div className="flex flex-wrap items-center gap-2">
      <Input
        className="w-full sm:w-64"
        placeholder={t("searchPlaceholder")}
        value={filters.search}
        onChange={handleSearch}
      />
      <FilterSidebar
        triggerLabel={tCommon("trigger")}
        title={tCommon("title")}
        resetLabel={tCommon("reset")}
        activeCount={activeCount(filters)}
        onReset={handleReset}
      >
        <FilterSidebarSection label={t("statusLabel")}>
          <SelectCombobox
            multiple
            items={HTTP_STATUS_ITEMS}
            value={filters.resp_statuses.map(String)}
            onValueChange={(values) => onChange({ ...filters, resp_statuses: values.map(Number) })}
            placeholder={t("statusPlaceholder")}
            emptyLabel={t("statusEmpty")}
          />
        </FilterSidebarSection>

        <FilterSidebarSection label={t("dateLabel")}>
          <Popover open={dateOpen} onOpenChange={setDateOpen}>
            <PopoverTrigger asChild>
              <Button
                variant="outline"
                size="sm"
                className="w-full justify-start font-normal"
                data-empty={!filters.date_from && !filters.date_to}
              >
                <CalendarIcon className="size-4 shrink-0" />
                <span className="truncate text-xs">{buildDateLabel(filters, t)}</span>
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="start">
              <Calendar
                mode="range"
                selected={dateRange}
                onSelect={handleDateSelect}
                numberOfMonths={1}
              />
              {(filters.date_from || filters.date_to) && (
                <div className="border-t px-3 py-2">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full text-xs"
                    onClick={() => {
                      setDateRange(undefined);
                      onChange({ ...filters, date_from: "", date_to: "" });
                    }}
                  >
                    {t("clearDate")}
                  </Button>
                </div>
              )}
            </PopoverContent>
          </Popover>
        </FilterSidebarSection>
      </FilterSidebar>
    </div>
  );
}
