"use client";

import { useCallback, useState } from "react";
import { fieldErrorsFromBody, type ApiBody, type FieldErrors } from "@/lib/api";
import { toast } from "sonner";

type SubmitArgs = {
  request: () => Promise<{ ok: boolean; status: number; body: ApiBody }>;
  successMessage: string;
  onSuccess?: () => void;
  onClose: (open: boolean) => void;
};

export function useDialogSubmit(genericError: string) {
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [submitting, setSubmitting] = useState(false);

  const submit = useCallback(
    async ({ request, successMessage, onSuccess, onClose }: SubmitArgs) => {
      setFieldErrors({});
      setSubmitting(true);
      try {
        const { ok, status, body } = await request();
        if (ok && body.success) {
          toast.success(successMessage);
          onSuccess?.();
          onClose(false);
          return;
        }
        if (status === 422 && Array.isArray(body.field_errors)) {
          setFieldErrors(fieldErrorsFromBody(body));
          return;
        }
        toast.error(body.message ?? genericError);
      } catch {
        toast.error(genericError);
      } finally {
        setSubmitting(false);
      }
    },
    [genericError],
  );

  return { fieldErrors, setFieldErrors, submitting, submit };
}
