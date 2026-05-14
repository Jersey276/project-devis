import { Suspense } from "react";
import LoginForm from "@/components/auth/login-form";

export default function loginPage() {
  return (
    <Suspense>
      <LoginForm />
    </Suspense>
  );
}
