"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Loader2Icon } from "lucide-react";
import { listTemplates } from "@/lib/services/templates";
import type { BackendTemplate } from "@/types/backend";

type SelectQuoteTemplateDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSelect: (templateId: string) => Promise<void>;
};

export default function SelectQuoteTemplateDialog({
  open,
  onOpenChange,
  onSelect,
}: SelectQuoteTemplateDialogProps) {
  const t = useTranslations("template.selectQuoteDialog");
  const [templates, setTemplates] = useState<BackendTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectingId, setSelectingId] = useState<string | null>(null);

  function handleOpenChange(newOpen: boolean) {
    if (!newOpen) setLoading(true);
    onOpenChange(newOpen);
  }

  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    listTemplates({ type: "quote_document" }).then(({ ok, body }) => {
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
      onOpenChange(false);
    } finally {
      setSelectingId(null);
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("title")}</DialogTitle>
          <DialogDescription>{t("description")}</DialogDescription>
        </DialogHeader>
        {loading ? (
          <div className="flex justify-center py-6">
            <Loader2Icon className="text-muted-foreground size-6 animate-spin" />
          </div>
        ) : templates.length === 0 ? (
          <p className="text-muted-foreground py-4 text-center text-sm">
            {t("empty")}
          </p>
        ) : (
          <div className="flex flex-col gap-2">
            {templates.map((tpl) => (
              <Button
                key={tpl.template_id}
                variant="outline"
                className="justify-start"
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
      </DialogContent>
    </Dialog>
  );
}
