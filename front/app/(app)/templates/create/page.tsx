import { getTranslations } from "next-intl/server";
import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default async function CreateTemplatePage() {
  const t = await getTranslations("templates");
  const breadcrumbs = [
    { href: "/templates", label: t("breadcrumb.list") },
    { href: "/templates/create", label: t("breadcrumb.create") },
  ];
  return (
    <>
      <PageBreadcrumb items={breadcrumbs} />
      <Card>
        <CardHeader>
          <CardTitle>{t("create.title")}</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-sm">{t("create.coming_soon")}</p>
        </CardContent>
      </Card>
    </>
  );
}
