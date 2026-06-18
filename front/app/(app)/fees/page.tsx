import { getTranslations } from "next-intl/server";
import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import SubscriptionGuard from "@/components/custom/subscription-guard";
import FeesTable from "@/components/fees/fees-table";

export default async function FeesPage() {
  const t = await getTranslations("fees");
  const breadcrumbs = [
    {
      href: "/fees",
      label: t("breadcrumb.list"),
    },
  ];
  return (
    <>
      <PageBreadcrumb items={breadcrumbs} />
      <Card>
        <CardHeader>
          <CardTitle>{t("list.title")}</CardTitle>
        </CardHeader>
        <CardContent>
          <SubscriptionGuard>
            <FeesTable />
          </SubscriptionGuard>
        </CardContent>
      </Card>
    </>
  );
}
