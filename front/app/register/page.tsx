import { FormEvent } from "react";
import { AuthLayout } from "../layout";
import LoginForm from "@/components/auth/login-form";
import RegisterForm from "@/components/auth/register-form";

export default function loginPage() {
  return (
    <AuthLayout>
      <RegisterForm />
    </AuthLayout>
  );
}
