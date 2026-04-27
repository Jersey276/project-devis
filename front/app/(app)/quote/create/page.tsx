import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import QuoteForm from "@/components/quote/quote-form";

const breadcrumbs = [
  {
    href: "/quote",
    label: "Devis",
  },
  {
    href: "/quote/create",
    label: "Nouveau devis",
  },
];

export default function CreateQuotePage() {
  return (
    <>
      <PageBreadcrumb items={breadcrumbs} />
      <QuoteForm />
    </>
  );
}
