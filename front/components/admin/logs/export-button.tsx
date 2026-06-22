"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { apiFetch } from "@/lib/api";
import type { LogFilters } from "./logs-filters";

type ExportButtonProps = {
  filters: LogFilters;
};

export default function ExportButton({ filters }: ExportButtonProps) {
  const t = useTranslations("admin.logs");
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ ok: boolean; text: string } | null>(null);

  const handleExport = async () => {
    setLoading(true);
    setMessage(null);

    const { ok } = await apiFetch("/api/logs/export", {
      method: "POST",
      body: JSON.stringify({
        filters: {
          user_id: filters.user_id,
          url_contains: filters.url_contains,
          resp_status: filters.resp_status ? Number(filters.resp_status) : 0,
          date_from: filters.date_from,
          date_to: filters.date_to,
        },
      }),
    });

    setMessage({ ok, text: ok ? t("exportSuccess") : t("exportError") });
    setLoading(false);
  };

  return (
    <div className="flex flex-col items-end gap-1">
      <Button variant="outline" size="sm" disabled={loading} onClick={handleExport}>
        {loading ? "Envoi…" : t("exportButton")}
      </Button>
      {message && (
        <p className={`text-xs ${message.ok ? "text-green-600" : "text-destructive"}`}>
          {message.text}
        </p>
      )}
    </div>
  );
}
