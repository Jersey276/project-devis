"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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
  const [search, setSearch] = useState("");
  const [visibleCount, setVisibleCount] = useState(15);
  const sentinelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    setLoading(true);
    setSearch("");
    setVisibleCount(15);
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

  const filtered = useMemo(() => {
    const list =
      search.trim() === ""
        ? templates
        : templates.filter((tpl) =>
            tpl.name.toLowerCase().includes(search.toLowerCase()),
          );
    return list;
  }, [templates, search]);

  useEffect(() => {
    setVisibleCount(15);
  }, [search]);

  const visible = filtered.slice(0, visibleCount);
  const hasMore = visibleCount < filtered.length;

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel || !hasMore) return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) setVisibleCount((c) => c + 15);
      },
      { threshold: 0.1 },
    );
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [hasMore, visible.length]);

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
                {visible.map((tpl) => (
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
                {hasMore && <div ref={sentinelRef} className="py-1" />}
              </div>
            )}
          </div>
        )}
      </PopoverContent>
    </Popover>
  );
}
