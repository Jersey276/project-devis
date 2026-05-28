"use client";

import { useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { EllipsisVerticalIcon, Loader2Icon } from "lucide-react";
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
import { Button } from "@/components/ui/button";
import { listTemplates } from "@/lib/services/templates";
import { listAvailableTaxesForUser } from "@/lib/services/taxes";
import type { BackendTax, BackendTemplate } from "@/types/backend";
import EditLineTemplateDialog from "@/components/template/edit-line-template-dialog";
import EditQuoteTemplateDialog from "@/components/template/edit-quote-template-dialog";

export default function TemplateTabs() {
  const t = useTranslations("templates");

  const [lineTemplates, setLineTemplates] = useState<BackendTemplate[]>([]);
  const [quoteTemplates, setQuoteTemplates] = useState<BackendTemplate[]>([]);
  const [loadingLines, setLoadingLines] = useState(true);
  const [loadingQuotes, setLoadingQuotes] = useState(true);
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
    listTemplates({ type: "quote_line" }).then(({ ok, body }) => {
      setLoadingLines(false);
      if (ok && Array.isArray(body.templates)) {
        setLineTemplates(body.templates as BackendTemplate[]);
      } else {
        toast.error(t("list.loadFailedToast"));
      }
    });
  }, [t]);

  useEffect(() => {
    listTemplates({ type: "quote_document" }).then(({ ok, body }) => {
      setLoadingQuotes(false);
      if (ok && Array.isArray(body.templates)) {
        setQuoteTemplates(body.templates as BackendTemplate[]);
      } else {
        toast.error(t("list.loadFailedToast"));
      }
    });
  }, [t]);

  function handleLineSaved(updated: BackendTemplate) {
    setLineTemplates((prev) =>
      prev.map((tpl) =>
        tpl.template_id === updated.template_id ? updated : tpl,
      ),
    );
  }

  function handleQuoteSaved(updated: BackendTemplate) {
    setQuoteTemplates((prev) =>
      prev.map((tpl) =>
        tpl.template_id === updated.template_id ? updated : tpl,
      ),
    );
  }

  return (
    <>
      <Tabs defaultValue="quote_line" className="mt-2">
        <TabsList>
          <TabsTrigger value="quote_line">{t("tabs.lines")}</TabsTrigger>
          <TabsTrigger value="quote_document">{t("tabs.quotes")}</TabsTrigger>
        </TabsList>

        <TabsContent value="quote_line" className="mt-4">
          <TemplateTable
            templates={lineTemplates}
            loading={loadingLines}
            onEdit={(id) => setEditLineId(id)}
          />
        </TabsContent>

        <TabsContent value="quote_document" className="mt-4">
          <TemplateTable
            templates={quoteTemplates}
            loading={loadingQuotes}
            onEdit={(id) => setEditQuoteId(id)}
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
};

function TemplateTable({ templates, loading, onEdit }: TemplateTableProps) {
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
        {templates.map((tpl) => (
          <TableRow key={tpl.template_id}>
            <TableCell className="font-medium">{tpl.name}</TableCell>
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
                  <DropdownMenuItem onClick={() => onEdit(tpl.template_id)}>
                    {t("list.actions.edit")}
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
