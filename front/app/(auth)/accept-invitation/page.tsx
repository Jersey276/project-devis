import { Suspense } from "react";
import AcceptInvitationForm from "@/components/auth/accept-invitation-form";

export default function AcceptInvitationPage() {
  return (
    <Suspense>
      <AcceptInvitationForm />
    </Suspense>
  );
}
