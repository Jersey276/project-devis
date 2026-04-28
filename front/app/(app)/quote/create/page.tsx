import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import QuoteForm from "@/components/quote/quote-form";
import CustomerRedirect from "@/components/quote/customer-redirect";

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
      <CustomerRedirect to="/quote" />
      <PageBreadcrumb items={breadcrumbs} />
      <QuoteForm />
    </>
  );
}
