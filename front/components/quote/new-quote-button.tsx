"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ChevronDownIcon } from "lucide-react";
import { useMode } from "@/lib/mode-context";
import SelectQuoteTemplateDialog from "@/components/template/select-quote-template-dialog";

export default function NewQuoteButton() {
  const { isCustomer } = useMode();
  const t = useTranslations("quote.list");
  const router = useRouter();
  const [templateDialogOpen, setTemplateDialogOpen] = useState(false);

  if (isCustomer) return null;

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button>
            {t("newButton")}
            <ChevronDownIcon className="ml-1 size-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem asChild>
            <Link href="/quote/create">{t("newBlank")}</Link>
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => setTemplateDialogOpen(true)}>
            {t("newFromTemplate")}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <SelectQuoteTemplateDialog
        open={templateDialogOpen}
        onOpenChange={setTemplateDialogOpen}
        onSelect={async (templateId) => {
          router.push(
            `/quote/create?template=${encodeURIComponent(templateId)}`,
          );
        }}
      />
    </>
  );
}
