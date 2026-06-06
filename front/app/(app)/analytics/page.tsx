import { getTranslations } from "next-intl/server";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import AdminGuard from "@/components/custom/admin-guard";
import AnalyticsDashboard from "@/components/admin/analytics/analytics-dashboard";

export default async function AnalyticsPage() {
  const t = await getTranslations("admin.analytics");

  return (
    <AdminGuard>
      <Card>
        <CardHeader>
          <CardTitle>{t("title")}</CardTitle>
        </CardHeader>
        <CardContent>
          <AnalyticsDashboard />
        </CardContent>
      </Card>
    </AdminGuard>
  );
}
