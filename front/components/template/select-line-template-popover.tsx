"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Button } from "@/components/ui/button";
import { Loader2Icon } from "lucide-react";
import { listTemplates } from "@/lib/services/templates";
import type { BackendTemplate } from "@/types/backend";

type SelectLineTemplatePopoverProps = {
  children: React.ReactNode;
  disabled?: boolean;
  onSelect: (templateId: string) => Promise<void>;
};

export default function SelectLineTemplatePopover({
  children,
  disabled,
  onSelect,
}: SelectLineTemplatePopoverProps) {
  const t = useTranslations("template.selectLinePopover");
  const [open, setOpen] = useState(false);
  const [templates, setTemplates] = useState<BackendTemplate[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectingId, setSelectingId] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    setLoading(true);
    listTemplates({ type: "quote_line" }).then(({ ok, body }) => {
      if (cancelled) return;
      setLoading(false);
      if (ok && Array.isArray(body.templates)) {
        setTemplates(body.templates as BackendTemplate[]);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [open]);

  async function handleSelect(templateId: string) {
    setSelectingId(templateId);
    try {
      await onSelect(templateId);
      setOpen(false);
    } finally {
      setSelectingId(null);
    }
  }

  return (
    <Popover open={open} onOpenChange={disabled ? undefined : setOpen}>
      <PopoverTrigger asChild>{children}</PopoverTrigger>
      <PopoverContent className="w-64 p-2">
        {loading ? (
          <div className="flex justify-center py-4">
            <Loader2Icon className="text-muted-foreground size-5 animate-spin" />
          </div>
        ) : templates.length === 0 ? (
          <p className="text-muted-foreground py-2 text-center text-sm">
            {t("empty")}
          </p>
        ) : (
          <div className="flex flex-col gap-1">
            {templates.map((tpl) => (
              <Button
                key={tpl.template_id}
                variant="ghost"
                className="w-full justify-start"
                disabled={selectingId !== null}
                onClick={() => handleSelect(tpl.template_id)}
              >
                {selectingId === tpl.template_id && (
                  <Loader2Icon className="mr-2 size-4 animate-spin" />
                )}
                {tpl.name}
              </Button>
            ))}
          </div>
        )}
      </PopoverContent>
    </Popover>
  );
}
