import { getTranslations } from "next-intl/server";
import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import TemplateTabs from "@/components/template/template-tabs";
import SubscriptionGuard from "@/components/custom/subscription-guard";

export default async function TemplatesPage() {
  const t = await getTranslations("templates");
  const breadcrumbs = [
    {
      href: "/templates",
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
            <TemplateTabs />
          </SubscriptionGuard>
        </CardContent>
      </Card>
    </>
  );
}
