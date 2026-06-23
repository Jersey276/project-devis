import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import AdminGuard from "@/components/custom/admin-guard";
import LogsDashboard from "@/components/admin/logs/logs-dashboard";

export default async function LogsPage() {
  const t = await getTranslations("admin.logs");

  return (
    <AdminGuard>
      <Card>
        <CardHeader>
          <CardTitle>{t("title")}</CardTitle>
        </CardHeader>
        <CardContent>
          <Suspense fallback={null}>
            <LogsDashboard />
          </Suspense>
        </CardContent>
      </Card>
    </AdminGuard>
  );
}
