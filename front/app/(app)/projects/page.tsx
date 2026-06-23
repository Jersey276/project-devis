import { getTranslations } from "next-intl/server";
import { Suspense } from "react";
import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import { Card, CardContent, CardHeader, CardTitle, CardAction } from "@/components/ui/card";
import ProjectListTable from "@/components/project/project-list-table";
import CreateProjectDialog from "@/components/project/create-project-dialog";

export default async function ProjectsIndexPage() {
  const t = await getTranslations("project");
  const breadcrumbs = [{ href: "/projects", label: t("list.title") }];

  return (
    <>
      <PageBreadcrumb items={breadcrumbs} />
      <Card>
        <CardHeader>
          <CardTitle>{t("list.title")}</CardTitle>
          <CardAction>
            <CreateProjectDialog />
          </CardAction>
        </CardHeader>
        <CardContent>
          <Suspense>
            <ProjectListTable />
          </Suspense>
        </CardContent>
      </Card>
    </>
  );
}
