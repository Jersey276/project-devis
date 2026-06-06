import { getTranslations } from "next-intl/server";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import AdminGuard from "@/components/custom/admin-guard";
import SubscriptionsTable from "@/components/admin/subscriptions/subscriptions-table";
import PlansTable from "@/components/admin/subscriptions/plans-table";

export default async function SubscriptionsPage() {
  const [tSubs, tPlans] = await Promise.all([
    getTranslations("admin.subscriptions"),
    getTranslations("admin.plans"),
  ]);

  return (
    <AdminGuard>
      <div className="grid gap-6">
        <Card>
          <CardHeader>
            <CardTitle>{tPlans("title")}</CardTitle>
          </CardHeader>
          <CardContent>
            <PlansTable />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>{tSubs("title")}</CardTitle>
          </CardHeader>
          <CardContent>
            <SubscriptionsTable />
          </CardContent>
        </Card>
      </div>
    </AdminGuard>
  );
}
