import { AppLayout } from "@/app/layout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import TaxesTable from "@/components/admin/taxes/taxes-table";

export default function TaxesPage() {
  return (
    <AppLayout>
      <Card>
        <CardHeader>
          <CardTitle>Taxes</CardTitle>
        </CardHeader>
        <CardContent>
          <TaxesTable />
        </CardContent>
      </Card>
    </AppLayout>
  );
}
