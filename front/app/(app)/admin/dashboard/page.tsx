import { getTranslations } from "next-intl/server";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import AdminGuard from "@/components/custom/admin-guard";
import AdminDashboard from "@/components/dashboard/admin-dashboard";

export default async function AdminDashboardPage() {
  const t = await getTranslations("nav");

  return (
    <AdminGuard>
      <Card>
        <CardHeader>
          <CardTitle>{t("adminDashboard")}</CardTitle>
        </CardHeader>
        <CardContent>
          <AdminDashboard />
        </CardContent>
      </Card>
    </AdminGuard>
  );
}
