"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import ClientForm, {
  EMPTY_CLIENT_VALUES,
  type ClientFormValues,
} from "@/components/user/client/client-form";
import { fieldErrorsFromBody, type FieldErrors } from "@/lib/api";
import { getClient, updateClient } from "@/lib/services/clients";
import type { BackendClient } from "@/types/backend";

function clientFromBackend(c: BackendClient): ClientFormValues {
  return {
    first_name: c.first_name,
    last_name: c.last_name,
    email: c.email ?? "",
    phone: c.phone ?? "",
    company: c.company ?? "",
    siren: c.siren ?? "",
    siret: c.siret ?? "",
    vat: c.vat ?? "",
    client_type: c.client_type || "individual",
  };
}

export default function EditClientPage() {
  const router = useRouter();
  const { uuid } = useParams<{ uuid: string }>();
  const t = useTranslations("client.edit");
  const tCommon = useTranslations("common");
  const [client, setClient] = useState<ClientFormValues>(EMPTY_CLIENT_VALUES);
  const [isLinked, setIsLinked] = useState(false);
  const [loaded, setLoaded] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});

  useEffect(() => {
    let cancelled = false;
    getClient(uuid).then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && body.success) {
        const c = body.client as BackendClient;
        setClient(clientFromBackend(c));
        setIsLinked(!!c.linked_user_id);
        setLoaded(true);
      } else {
        toast.error((body.message as string) ?? t("notFoundToast"));
        router.push("/clients");
      }
    });
    return () => {
      cancelled = true;
    };
  }, [uuid, router, t]);

  async function handleSubmit() {
    if (submitting || isLinked) return;
    setFieldErrors({});
    setSubmitting(true);
    try {
      const { ok, body } = await updateClient(uuid, client);
      if (ok && body.success) {
        toast.success(t("successToast"));
        router.push(`/clients/${uuid}`);
        return;
      }
      const parsed = fieldErrorsFromBody(body);
      if (Object.keys(parsed).length > 0) {
        setFieldErrors(parsed);
      } else {
        toast.error((body.message as string) ?? t("failedToast"));
      }
    } finally {
      setSubmitting(false);
    }
  }

  if (!loaded) {
    return (
      <Card>
        <CardContent className="py-16 text-center text-muted-foreground">
          {tCommon("actions.loading")}
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="max-w-3xl">
      <CardHeader>
        <CardTitle>{t("title")}</CardTitle>
        {isLinked ? (
          <CardDescription className="text-amber-600">
            {t("linkedNotice")}
          </CardDescription>
        ) : (
          <CardDescription>{t("description")}</CardDescription>
        )}
      </CardHeader>

      <CardContent>
        <ClientForm
          client={client}
          onClientChange={isLinked ? () => {} : setClient}
          fieldErrors={isLinked ? {} : fieldErrors}
        />
      </CardContent>

      <CardFooter className="justify-end gap-2">
        <Button
          variant="outline"
          type="button"
          onClick={() => router.push(`/clients/${uuid}`)}
        >
          {isLinked ? tCommon("actions.close") : tCommon("actions.cancel")}
        </Button>
        {!isLinked && (
          <Button type="button" onClick={handleSubmit} disabled={submitting}>
            {submitting ? tCommon("actions.saving") : tCommon("actions.save")}
          </Button>
        )}
      </CardFooter>
    </Card>
  );
}
