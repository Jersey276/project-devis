"use client";

import { useTranslations } from "next-intl";
import { Input } from "@/components/ui/input";
import { SelectCombobox } from "@/components/ui/select-combobox";
import { FilterSidebar, FilterSidebarSection } from "@/components/ui/filter-sidebar";
import { DateRangePicker } from "@/components/ui/date-range-picker";

export type EmailLogFilters = {
  search: string;
  statuses: string[];
  types: string[];
  date_from: string;
  date_to: string;
};

export const emptyEmailLogFilters: EmailLogFilters = {
  search: "",
  statuses: [],
  types: [],
  date_from: "",
  date_to: "",
};

const STATUS_ITEMS = [
  { value: "sent", label: "Envoyé" },
  { value: "failed", label: "Échoué" },
];

const TYPE_ITEMS = [
  { value: "quote_sent", label: "Devis envoyé" },
  { value: "schedule_VALID", label: "Échéancier validé" },
  { value: "schedule_DENIED", label: "Échéancier refusé" },
  { value: "generic", label: "Générique" },
];

type Props = {
  filters: EmailLogFilters;
  onChange: (filters: EmailLogFilters) => void;
};

function activeCount(f: EmailLogFilters): number {
  return (
    (f.statuses.length > 0 ? 1 : 0) +
    (f.types.length > 0 ? 1 : 0) +
    (f.date_from || f.date_to ? 1 : 0)
  );
}

export default function EmailLogsFilters({ filters, onChange }: Props) {
  const t = useTranslations("admin.emailLogs.filters");
  const tCommon = useTranslations("common.filterSidebar");

  return (
    <div className="flex flex-wrap items-center gap-2">
      <Input
        className="w-full sm:w-64"
        placeholder={t("searchPlaceholder")}
        value={filters.search}
        onChange={(e) => onChange({ ...filters, search: e.target.value })}
      />
      <FilterSidebar
        triggerLabel={tCommon("trigger")}
        title={tCommon("title")}
        resetLabel={tCommon("reset")}
        activeCount={activeCount(filters)}
        onReset={() => onChange(emptyEmailLogFilters)}
      >
        <FilterSidebarSection label={t("statusLabel")}>
          <SelectCombobox
            multiple
            items={STATUS_ITEMS}
            value={filters.statuses}
            onValueChange={(values) => onChange({ ...filters, statuses: values })}
            placeholder={t("statusPlaceholder")}
            emptyLabel={t("statusEmpty")}
          />
        </FilterSidebarSection>

        <FilterSidebarSection label={t("typeLabel")}>
          <SelectCombobox
            multiple
            items={TYPE_ITEMS}
            value={filters.types}
            onValueChange={(values) => onChange({ ...filters, types: values })}
            placeholder={t("typePlaceholder")}
            emptyLabel={t("typeEmpty")}
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
