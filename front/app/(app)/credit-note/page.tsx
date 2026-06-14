import { getTranslations } from "next-intl/server";
import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import SubscriptionGuard from "@/components/custom/subscription-guard";
import CreditNoteListTable from "@/components/invoice/credit-note-list-table";

export default async function CreditNoteIndexPage() {
  const t = await getTranslations("creditNote.list");
  const breadcrumbs = [{ href: "/credit-note", label: t("title") }];

  return (
    <>
      <PageBreadcrumb items={breadcrumbs} />
      <Card>
        <CardHeader>
          <CardTitle>{t("title")}</CardTitle>
        </CardHeader>
        <CardContent>
          <SubscriptionGuard>
            <CreditNoteListTable />
          </SubscriptionGuard>
        </CardContent>
      </Card>
    </>
  );
}
