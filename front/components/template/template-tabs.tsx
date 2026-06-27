"use client";

import { useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { ArchiveIcon, EllipsisVerticalIcon, Loader2Icon, RotateCcwIcon } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { listTemplates, archiveTemplate, restoreTemplate } from "@/lib/services/templates";
import { listAvailableTaxesForUser } from "@/lib/services/taxes";
import type { BackendTax, BackendTemplate } from "@/types/backend";
import EditLineTemplateDialog from "@/components/template/edit-line-template-dialog";
import EditQuoteTemplateDialog from "@/components/template/edit-quote-template-dialog";

export default function TemplateTabs() {
  const t = useTranslations("templates");

  const [lineTemplates, setLineTemplates] = useState<BackendTemplate[] | null>(null);
  const [quoteTemplates, setQuoteTemplates] = useState<BackendTemplate[] | null>(null);
  const [showArchived, setShowArchived] = useState(false);
  const [availableTaxes, setAvailableTaxes] = useState<BackendTax[]>([]);

  const [editLineId, setEditLineId] = useState<string | null>(null);
  const [editQuoteId, setEditQuoteId] = useState<string | null>(null);

  const taxById = useMemo(
    () => new Map(availableTaxes.map((tax) => [tax.id, tax])),
    [availableTaxes],
  );

  useEffect(() => {
    listAvailableTaxesForUser().then(({ ok, body }) => {
      if (ok && Array.isArray(body.taxes)) {
        setAvailableTaxes(body.taxes as BackendTax[]);
      }
    });
  }, []);

  useEffect(() => {
    let cancelled = false;
    listTemplates({ type: "quote_line", archived: showArchived }).then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && Array.isArray(body.templates)) {
        setLineTemplates(body.templates as BackendTemplate[]);
      } else {
        setLineTemplates([]);
        toast.error(t("list.loadFailedToast"));
      }
    });
    return () => { cancelled = true; };
  }, [t, showArchived]);

  useEffect(() => {
    let cancelled = false;
    listTemplates({ type: "quote_document", archived: showArchived }).then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && Array.isArray(body.templates)) {
        setQuoteTemplates(body.templates as BackendTemplate[]);
      } else {
        setQuoteTemplates([]);
        toast.error(t("list.loadFailedToast"));
      }
    });
    return () => { cancelled = true; };
  }, [t, showArchived]);

  function handleLineSaved(updated: BackendTemplate) {
    setLineTemplates((prev) =>
      prev?.map((tpl) => (tpl.template_id === updated.template_id ? updated : tpl)) ?? prev,
    );
  }

  function handleQuoteSaved(updated: BackendTemplate) {
    setQuoteTemplates((prev) =>
      prev?.map((tpl) => (tpl.template_id === updated.template_id ? updated : tpl)) ?? prev,
    );
  }

  async function handleArchive(id: string, setter: React.Dispatch<React.SetStateAction<BackendTemplate[] | null>>) {
    const { ok, body } = await archiveTemplate(id);
    if (ok && body.success) {
      toast.success(t("list.archiveSuccessToast"));
      setter((prev) => prev?.filter((tpl) => tpl.template_id !== id) ?? prev);
    } else {
      toast.error((body.message as string) ?? t("list.archiveFailedToast"));
    }
  }

  async function handleRestore(id: string, setter: React.Dispatch<React.SetStateAction<BackendTemplate[] | null>>) {
    const { ok, body } = await restoreTemplate(id);
    if (ok && body.success) {
      toast.success(t("list.restoreSuccessToast"));
      setter((prev) => prev?.filter((tpl) => tpl.template_id !== id) ?? prev);
    } else {
      toast.error((body.message as string) ?? t("list.restoreFailedToast"));
    }
  }

  return (
    <>
      <div className="flex items-center gap-2 mt-2 mb-4">
        <Checkbox
          id="template-archived"
          checked={showArchived}
          onCheckedChange={(checked) => {
            setLineTemplates(null);
            setQuoteTemplates(null);
            setShowArchived(!!checked);
          }}
        />
        <Label htmlFor="template-archived">{t("list.showArchivedLabel")}</Label>
      </div>

      <Tabs defaultValue="quote_line">
        <TabsList>
          <TabsTrigger value="quote_line">{t("tabs.lines")}</TabsTrigger>
          <TabsTrigger value="quote_document">{t("tabs.quotes")}</TabsTrigger>
        </TabsList>

        <TabsContent value="quote_line" className="mt-4">
          <TemplateTable
            templates={lineTemplates ?? []}
            loading={lineTemplates === null}
            onEdit={(id) => setEditLineId(id)}
            onArchive={(id) => handleArchive(id, setLineTemplates)}
            onRestore={(id) => handleRestore(id, setLineTemplates)}
          />
        </TabsContent>

        <TabsContent value="quote_document" className="mt-4">
          <TemplateTable
            templates={quoteTemplates ?? []}
            loading={quoteTemplates === null}
            onEdit={(id) => setEditQuoteId(id)}
            onArchive={(id) => handleArchive(id, setQuoteTemplates)}
            onRestore={(id) => handleRestore(id, setQuoteTemplates)}
          />
        </TabsContent>
      </Tabs>

      {editLineId !== null && (
        <EditLineTemplateDialog
          templateId={editLineId}
          open
          onOpenChange={(o) => {
            if (!o) setEditLineId(null);
          }}
          availableTaxes={availableTaxes}
          taxById={taxById}
          onSaved={handleLineSaved}
        />
      )}

      {editQuoteId !== null && (
        <EditQuoteTemplateDialog
          templateId={editQuoteId}
          open
          onOpenChange={(o) => {
            if (!o) setEditQuoteId(null);
          }}
          availableTaxes={availableTaxes}
          taxById={taxById}
          onSaved={handleQuoteSaved}
        />
      )}
    </>
  );
}

type TemplateTableProps = {
  templates: BackendTemplate[];
  loading: boolean;
  onEdit: (id: string) => void;
  onArchive: (id: string) => void;
  onRestore: (id: string) => void;
};

function TemplateTable({ templates, loading, onEdit, onArchive, onRestore }: TemplateTableProps) {
  const t = useTranslations("templates");
  if (loading) {
    return (
      <div className="flex justify-center py-10">
        <Loader2Icon className="text-muted-foreground size-5 animate-spin" />
      </div>
    );
  }

  if (templates.length === 0) {
    return (
      <p className="text-muted-foreground py-6 text-center text-sm">
        {t("list.empty")}
      </p>
    );
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>{t("list.columns.name")}</TableHead>
          <TableHead>{t("list.columns.createdAt")}</TableHead>
          <TableHead className="w-16">{t("list.columns.actions")}</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {templates.map((tpl) => {
          const isArchived = !!tpl.archived_at;
          return (
            <TableRow key={tpl.template_id} className={isArchived ? "opacity-60" : undefined}>
              <TableCell className="font-medium">
                <span className="flex items-center gap-2">
                  {tpl.name}
                  {isArchived && (
                    <Badge variant="secondary" className="gap-1 text-xs">
                      <ArchiveIcon className="size-3" />
                      {t("list.archivedBadge")}
                    </Badge>
                  )}
                </span>
              </TableCell>
              <TableCell className="text-muted-foreground text-sm">
                {new Date(tpl.created_at).toLocaleDateString("fr-FR")}
              </TableCell>
              <TableCell>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label={t("list.columns.actions")}
                    >
                      <EllipsisVerticalIcon className="size-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    {!isArchived && (
                      <DropdownMenuItem onClick={() => onEdit(tpl.template_id)}>
                        {t("list.actions.edit")}
                      </DropdownMenuItem>
                    )}
                    {!isArchived ? (
                      <DropdownMenuItem onClick={() => onArchive(tpl.template_id)}>
                        <ArchiveIcon className="size-4" />
                        {t("list.actions.archive")}
                      </DropdownMenuItem>
                    ) : (
                      <DropdownMenuItem onClick={() => onRestore(tpl.template_id)}>
                        <RotateCcwIcon className="size-4" />
                        {t("list.actions.restore")}
                      </DropdownMenuItem>
                    )}
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          );
        })}
      </TableBody>
    </Table>
  );
}
