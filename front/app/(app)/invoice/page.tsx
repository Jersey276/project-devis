import { getTranslations } from "next-intl/server";
import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import SubscriptionGuard from "@/components/custom/subscription-guard";
import InvoiceListTable from "@/components/invoice/invoice-list-table";
import OSSThresholdBanner from "@/components/invoice/oss-threshold-banner";
import EReportingSection from "@/components/invoice/e-reporting-section";

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
            <OSSThresholdBanner />
            <InvoiceListTable />
            <div className="mt-8 border-t pt-6">
              <EReportingSection />
            </div>
          </SubscriptionGuard>
        </CardContent>
      </Card>
    </>
  );
}
