"use client";

import { useTranslations } from "next-intl";
import { Input } from "@/components/ui/input";
import { SelectCombobox } from "@/components/ui/select-combobox";
import { FilterSidebar, FilterSidebarSection } from "@/components/ui/filter-sidebar";
import { DateRangePicker } from "@/components/ui/date-range-picker";

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


type LogsFiltersProps = {
  filters: LogFilters;
  onChange: (filters: LogFilters) => void;
};

function activeCount(f: LogFilters): number {
  return (f.resp_statuses.length > 0 ? 1 : 0) + (f.date_from || f.date_to ? 1 : 0);
}

export default function LogsFilters({ filters, onChange }: LogsFiltersProps) {
  const t = useTranslations("admin.logs.filters");
  const tCommon = useTranslations("common.filterSidebar");

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...filters, search: e.target.value });
  };

  const handleReset = () => {
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
          <DateRangePicker
            from={filters.date_from}
            to={filters.date_to}
            onValueChange={(from, to) => onChange({ ...filters, date_from: from, date_to: to })}
          />
        </FilterSidebarSection>
      </FilterSidebar>
    </div>
  );
}
