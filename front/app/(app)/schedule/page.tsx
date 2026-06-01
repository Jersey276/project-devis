import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import ScheduleListTable from "@/components/schedule/schedule-list-table";

export default function ScheduleIndexPage() {
  const breadcrumbs = [{ href: "/schedule", label: "Échéanciers" }];

  return (
    <>
      <PageBreadcrumb items={breadcrumbs} />
      <Card>
        <CardHeader>
          <CardTitle>Échéanciers</CardTitle>
        </CardHeader>
        <CardContent>
          <ScheduleListTable />
        </CardContent>
      </Card>
    </>
  );
}
