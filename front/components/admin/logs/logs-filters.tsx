"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

export type LogFilters = {
  user_id: string;
  url_contains: string;
  resp_status: string;
  date_from: string;
  date_to: string;
};

type LogsFiltersProps = {
  filters: LogFilters;
  onChange: (filters: LogFilters) => void;
};

const empty: LogFilters = {
  user_id: "",
  url_contains: "",
  resp_status: "",
  date_from: "",
  date_to: "",
};

export default function LogsFilters({ filters, onChange }: LogsFiltersProps) {
  const t = useTranslations("admin.logs.filters");
  const [draft, setDraft] = useState<LogFilters>(filters);

  const set = (key: keyof LogFilters) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setDraft((prev) => ({ ...prev, [key]: e.target.value }));

  const apply = () => onChange(draft);
  const reset = () => {
    setDraft(empty);
    onChange(empty);
  };

  return (
    <div className="flex flex-wrap gap-2">
      <Input
        className="w-40"
        placeholder={t("userId")}
        value={draft.user_id}
        onChange={set("user_id")}
        onKeyDown={(e) => e.key === "Enter" && apply()}
      />
      <Input
        className="w-48"
        placeholder={t("urlContains")}
        value={draft.url_contains}
        onChange={set("url_contains")}
        onKeyDown={(e) => e.key === "Enter" && apply()}
      />
      <Input
        className="w-24"
        placeholder={t("respStatus")}
        value={draft.resp_status}
        onChange={set("resp_status")}
        onKeyDown={(e) => e.key === "Enter" && apply()}
      />
      <Input
        type="date"
        className="w-40"
        placeholder={t("dateFrom")}
        value={draft.date_from}
        onChange={set("date_from")}
      />
      <Input
        type="date"
        className="w-40"
        placeholder={t("dateTo")}
        value={draft.date_to}
        onChange={set("date_to")}
      />
      <Button variant="outline" size="sm" onClick={apply}>
        Filtrer
      </Button>
      <Button variant="ghost" size="sm" onClick={reset}>
        {t("reset")}
      </Button>
    </div>
  );
}
