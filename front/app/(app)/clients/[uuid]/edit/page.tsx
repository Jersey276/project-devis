"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
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
    vat: c.vat ?? "",
  };
}

export default function EditClientPage() {
  const router = useRouter();
  const { uuid } = useParams<{ uuid: string }>();
  const [client, setClient] = useState<ClientFormValues>(EMPTY_CLIENT_VALUES);
  const [loaded, setLoaded] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});

  useEffect(() => {
    let cancelled = false;
    getClient(uuid).then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && body.success) {
        setClient(clientFromBackend(body.client as BackendClient));
        setLoaded(true);
      } else {
        toast.error((body.message as string) ?? "Client introuvable.");
        router.push("/clients");
      }
    });
    return () => {
      cancelled = true;
    };
  }, [uuid, router]);

  async function handleSubmit() {
    if (submitting) return;
    setFieldErrors({});
    setSubmitting(true);
    try {
      const { ok, status, body } = await updateClient(uuid, client);
      if (ok && body.success) {
        toast.success("Client mis à jour.");
        router.push(`/clients/${uuid}`);
        return;
      }
      if (status === 422) {
        setFieldErrors(fieldErrorsFromBody(body));
      } else {
        toast.error(
          (body.message as string) ?? "Impossible de mettre à jour le client.",
        );
      }
    } finally {
      setSubmitting(false);
    }
  }

  if (!loaded) {
    return (
      <Card>
        <CardContent className="py-16 text-center text-muted-foreground">
          Chargement…
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="max-w-3xl">
      <CardHeader>
        <CardTitle>Modifier le client</CardTitle>
        <CardDescription>
          Mettez à jour les informations principales du client.
        </CardDescription>
      </CardHeader>

      <CardContent>
        <ClientForm
          client={client}
          onClientChange={setClient}
          fieldErrors={fieldErrors}
        />
      </CardContent>

      <CardFooter className="justify-end gap-2">
        <Button
          variant="outline"
          type="button"
          onClick={() => router.push(`/clients/${uuid}`)}
        >
          Annuler
        </Button>
        <Button type="button" onClick={handleSubmit} disabled={submitting}>
          {submitting ? "Enregistrement…" : "Enregistrer"}
        </Button>
      </CardFooter>
    </Card>
  );
}
