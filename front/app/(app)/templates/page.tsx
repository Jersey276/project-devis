import { getTranslations } from "next-intl/server";
import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import Link from "next/link";
import { Button } from "@/components/ui/button";

export default async function TemplatesPage() {
  const t = await getTranslations("templates");
  const breadcrumbs = [
    {
      href: "/templates",
      label: t("breadcrumb.list"),
    },
  ];
  return (
    <>
      <PageBreadcrumb items={breadcrumbs} />
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>{t("list.title")}</CardTitle>
          <Button asChild>
            <Link href="/templates/create">{t("list.new")}</Link>
          </Button>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-sm">{t("list.empty")}</p>
        </CardContent>
      </Card>
    </>
  );
}
