import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import AdminGuard from "@/components/custom/admin-guard";
import EmailLogsDashboard from "@/components/admin/email-logs/email-logs-dashboard";

export default async function EmailLogsPage() {
  const t = await getTranslations("admin.emailLogs");

  return (
    <AdminGuard>
      <Card>
        <CardHeader>
          <CardTitle>{t("title")}</CardTitle>
        </CardHeader>
        <CardContent>
          <Suspense fallback={null}>
            <EmailLogsDashboard />
          </Suspense>
        </CardContent>
      </Card>
    </AdminGuard>
  );
}
