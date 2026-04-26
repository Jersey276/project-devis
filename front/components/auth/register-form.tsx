"use client";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { useRouter } from "next/navigation";
import { toast } from "sonner";

// Must stay in sync with backend/auth/actions/errors.go field validation codes.
const FIELD_VALIDATION_MESSAGES: Record<number, string> = {
  1: "This field is required.",
  2: "Invalid format.",
  3: "Too short (minimum 8 characters).",
  4: "This email address is already in use.",
};

type FieldErrors = Record<string, string[]>;

function toMessages(codes: number[]): string[] {
  return codes.map(
    (code) => FIELD_VALIDATION_MESSAGES[code] ?? `Validation error (${code}).`,
  );
}

function toErrorProps(messages: string[] | undefined) {
  return messages?.map((message) => ({ message }));
}

export default function RegisterForm({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  const router = useRouter();
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [confirmError, setConfirmError] = useState<string | null>(null);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setFieldErrors({});
    setConfirmError(null);

    const form = e.currentTarget;
    const data = new FormData(form);
    const email = data.get("email") as string;
    const password = data.get("password") as string;
    const confirmPassword = data.get("confirm-password") as string;

    if (password !== confirmPassword) {
      setConfirmError("Passwords do not match.");
      return;
    }

    try {
      const response = await fetch("/api/auth/register", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify({ email, password }),
      });

      if (response.ok) {
        toast.success("Registration successful! Please log in.");
        router.replace("/login");
        return;
      }

      const body = await response.json();

      if (response.status === 422 && Array.isArray(body.field_errors)) {
        const errors: FieldErrors = {};
        for (const entry of body.field_errors as {
          field: string;
          error_code: number[];
        }[]) {
          errors[entry.field] = toMessages(entry.error_code);
        }
        setFieldErrors(errors);
        return;
      }

      toast.error("Registration failed. Please try again.");
    } catch {
      toast.error("Registration failed. Please try again.");
    }
  }

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader>
          <CardTitle>Create an account</CardTitle>
          <CardDescription>
            Enter your information below to create your account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} noValidate>
            <FieldGroup>
              <Field data-invalid={!!fieldErrors.email?.length}>
                <FieldLabel htmlFor="email">Email</FieldLabel>
                <Input
                  id="email"
                  type="email"
                  name="email"
                  placeholder="m@example.com"
                  aria-invalid={!!fieldErrors.email?.length}
                />
                <FieldError errors={toErrorProps(fieldErrors.email)} />
                <FieldDescription>
                  We&apos;ll use this to contact you. We will not share your
                  email with anyone else.
                </FieldDescription>
              </Field>
              <Field data-invalid={!!fieldErrors.password?.length}>
                <FieldLabel htmlFor="password">Password</FieldLabel>
                <Input
                  id="password"
                  type="password"
                  name="password"
                  aria-invalid={!!fieldErrors.password?.length}
                />
                <FieldError errors={toErrorProps(fieldErrors.password)} />
                <FieldDescription>
                  Must be at least 8 characters long.
                </FieldDescription>
              </Field>
              <Field data-invalid={!!confirmError}>
                <FieldLabel htmlFor="confirm-password">
                  Confirm Password
                </FieldLabel>
                <Input
                  id="confirm-password"
                  type="password"
                  name="confirm-password"
                  aria-invalid={!!confirmError}
                />
                <FieldError
                  errors={
                    confirmError ? [{ message: confirmError }] : undefined
                  }
                />
                <FieldDescription>
                  Please confirm your password.
                </FieldDescription>
              </Field>
              <FieldGroup>
                <Field>
                  <Button type="submit">Create Account</Button>
                  <Button variant="outline" type="button">
                    Sign up with Google
                  </Button>
                  <FieldDescription className="px-6 text-center">
                    Already have an account? <a href="/login">Sign in</a>
                  </FieldDescription>
                </Field>
              </FieldGroup>
            </FieldGroup>
          </form>
        </CardContent>
      </Card>
      <FieldDescription className="px-6 text-center">
        By clicking continue, you agree to our <a href="#">Terms of Service</a>{" "}
        and <a href="#">Privacy Policy</a>.
      </FieldDescription>
    </div>
  );
}
