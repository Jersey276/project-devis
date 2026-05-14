"use client";
import { FormEvent } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { NEXT_PARAM, safeNextPath } from "@/lib/auth-utils";
import { useRouter, useSearchParams } from "next/navigation";
import { toast } from "sonner";

function submitLoginForm(
  router: ReturnType<typeof useRouter>,
  next: string,
) {
  return async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const data = new FormData(form);
    const email = data.get("email");
    const password = data.get("password");
    const rememberMe = data.get("remember_me") === "on";
    await fetch("/api/auth/login", {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ email, password, remember_me: rememberMe }),
    })
      .then(async (response) => {
        if (response.ok) {
          toast.success("Login successful!");
          router.replace(next);
        } else {
          toast.error(
            "Login failed. Please check your credentials and try again.",
          );
        }
      })
      .catch(() => {
        toast.error(
          "Login failed. Please check your credentials and try again.",
        );
      });
  };
}

export default function LoginForm({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  const router = useRouter();
  const next = safeNextPath(useSearchParams().get(NEXT_PARAM));
  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader>
          <CardTitle>Login to your account</CardTitle>
          <CardDescription>
            Enter your email below to login to your account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={submitLoginForm(router, next)}>
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="email">Email</FieldLabel>
                <Input
                  id="email"
                  type="email"
                  name="email"
                  placeholder="m@example.com"
                  required
                />
              </Field>
              <Field>
                <div className="flex items-center">
                  <FieldLabel htmlFor="password">Password</FieldLabel>
                  <a
                    href="#"
                    className="ml-auto inline-block text-sm underline-offset-4 hover:underline"
                  >
                    Forgot your password?
                  </a>
                </div>
                <Input id="password" type="password" name="password" required />
              </Field>
              <Field>
                <div className="flex items-center gap-2">
                  <Checkbox id="remember_me" name="remember_me" />
                  <FieldLabel htmlFor="remember_me" className="font-normal">
                    Se souvenir de moi
                  </FieldLabel>
                </div>
              </Field>
              <Field>
                <Button type="submit">Login</Button>
                <Button variant="outline" type="button">
                  Login with Google
                </Button>
                <FieldDescription className="text-center">
                  Don&apos;t have an account? <a href="/register">Sign up</a>
                </FieldDescription>
              </Field>
            </FieldGroup>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
