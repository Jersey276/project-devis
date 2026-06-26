import { getTranslations } from "next-intl/server";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import DashboardRouter from "@/components/dashboard/dashboard-router";

export default async function DashboardPage() {
  const t = await getTranslations("dashboard");

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("title")}</CardTitle>
      </CardHeader>
      <CardContent>
        <DashboardRouter />
      </CardContent>
    </Card>
  );
}
