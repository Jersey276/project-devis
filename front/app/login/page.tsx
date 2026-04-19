import { FormEvent } from "react";
import { AuthLayout } from "../layout";
import LoginForm from "@/components/auth/login-form";

export default function loginPage() {
  return (
    <AuthLayout>
      <LoginForm />
    </AuthLayout>
  );
}
