import { getTranslations } from "next-intl/server";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import TaxesTable from "@/components/admin/taxes/taxes-table";
import AdminGuard from "@/components/custom/admin-guard";

export default async function TaxesPage() {
  const t = await getTranslations("admin.taxes");
  return (
    <AdminGuard>
      <Card>
        <CardHeader>
          <CardTitle>{t("title")}</CardTitle>
        </CardHeader>
        <CardContent>
          <TaxesTable />
        </CardContent>
      </Card>
    </AdminGuard>
  );
}
