import { getTranslations } from "next-intl/server";
import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import SubscriptionGuard from "@/components/custom/subscription-guard";
import InvoiceListTable from "@/components/invoice/invoice-list-table";

export default async function InvoiceIndexPage() {
  const t = await getTranslations("invoice.list");
  const breadcrumbs = [{ href: "/invoice", label: t("title") }];

  return (
    <>
      <PageBreadcrumb items={breadcrumbs} />
      <Card>
        <CardHeader>
          <CardTitle>{t("title")}</CardTitle>
        </CardHeader>
        <CardContent>
          <SubscriptionGuard>
            <InvoiceListTable />
          </SubscriptionGuard>
        </CardContent>
      </Card>
    </>
  );
}
