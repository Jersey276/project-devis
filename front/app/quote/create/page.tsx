import { AppLayout } from "@/app/layout";
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
    <AppLayout breadcrumbs={breadcrumbs}>
      <QuoteForm />
    </AppLayout>
  );
}
