import { getTranslations } from "next-intl/server";
import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import QuoteForm from "@/components/quote/quote-form";
import CustomerRedirect from "@/components/quote/customer-redirect";

export default async function CreateQuotePage() {
  const t = await getTranslations("quote.breadcrumb");
  const breadcrumbs = [
    {
      href: "/quote",
      label: t("list"),
    },
    {
      href: "/quote/create",
      label: t("create"),
    },
  ];
  return (
    <>
      <CustomerRedirect to="/quote" />
      <PageBreadcrumb items={breadcrumbs} />
      <QuoteForm />
    </>
  );
}
