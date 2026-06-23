import { getTranslations } from "next-intl/server";
import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import ProjectDetail from "@/components/project/project-detail";

type PageProps = { params: Promise<{ uuid: string }> };

export default async function ProjectDetailPage({ params }: PageProps) {
  const { uuid } = await params;
  const t = await getTranslations("project");
  const breadcrumbs = [
    { href: "/projects", label: t("list.title") },
    { href: `/projects/${uuid}`, label: t("detail.title") },
  ];

  return (
    <>
      <PageBreadcrumb items={breadcrumbs} />
      <ProjectDetail projectId={uuid} />
    </>
  );
}
