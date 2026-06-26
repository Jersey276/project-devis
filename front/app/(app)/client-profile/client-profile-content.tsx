"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import ClientForm, {
  EMPTY_CLIENT_VALUES,
  type ClientFormValues,
} from "@/components/user/client/client-form";
import AddressesTable from "@/components/address/addresses-table";
import { fieldErrorsFromBody, type FieldErrors } from "@/lib/api";
import {
  getMyClientProfiles,
  updateMyClientProfile,
} from "@/lib/services/clients";
import type { BackendClient } from "@/types/backend";

function clientFormValues(c: BackendClient): ClientFormValues {
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

function clientLabel(c: BackendClient): string {
  const name = [c.first_name, c.last_name].filter(Boolean).join(" ");
  return c.company ? `${c.company}${name ? ` (${name})` : ""}` : name || c.client_id;
}

export default function ClientProfileContent() {
  const t = useTranslations("clientProfile");
  const tCommon = useTranslations("common");

  const [clients, setClients] = useState<BackendClient[]>([]);
  const [selectedId, setSelectedId] = useState<string>("");
  const [form, setForm] = useState<ClientFormValues>(EMPTY_CLIENT_VALUES);
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [submitting, setSubmitting] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    getMyClientProfiles().then(({ ok, body }) => {
      if (cancelled) return;
      setLoading(false);
      if (ok && body.success && Array.isArray(body.clients) && body.clients.length > 0) {
        const list = body.clients as BackendClient[];
        setClients(list);
        setSelectedId(list[0].client_id);
        setForm(clientFormValues(list[0]));
      }
    });
    return () => { cancelled = true; };
  }, []);

  function handleSelectClient(clientId: string) {
    const c = clients.find((cl) => cl.client_id === clientId);
    if (!c) return;
    setSelectedId(clientId);
    setForm(clientFormValues(c));
    setFieldErrors({});
  }

  const selectedClient = clients.find((c) => c.client_id === selectedId) ?? null;

  async function handleSave() {
    if (submitting || !selectedId) return;
    setFieldErrors({});
    setSubmitting(true);
    try {
      const { ok, body } = await updateMyClientProfile(selectedId, form);
      if (ok && body.success) {
        toast.success(t("saveSuccess"));
      } else {
        const parsed = fieldErrorsFromBody(body);
        if (Object.keys(parsed).length > 0) {
          setFieldErrors(parsed);
        } else {
          toast.error((body.message as string) ?? t("saveError"));
        }
      }
    } finally {
      setSubmitting(false);
    }
  }

  if (loading) {
    return (
      <Card>
        <CardContent className="py-16 text-center text-muted-foreground">
          {tCommon("actions.loading")}
        </CardContent>
      </Card>
    );
  }

  if (clients.length === 0) {
    return (
      <Card>
        <CardContent className="py-16 text-center text-muted-foreground">
          {t("noLink")}
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="max-w-3xl">
      <CardHeader>
        <CardTitle>{t("title")}</CardTitle>
        {clients.length > 1 && (
          <Select value={selectedId} onValueChange={handleSelectClient}>
            <SelectTrigger className="mt-2 w-64">
              <SelectValue placeholder={t("providerPlaceholder")} />
            </SelectTrigger>
            <SelectContent>
              {clients.map((c) => (
                <SelectItem key={c.client_id} value={c.client_id}>
                  {clientLabel(c)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="information">
          <TabsList className="mb-4">
            <TabsTrigger value="information">{t("tabs.information")}</TabsTrigger>
            <TabsTrigger value="addresses">{t("tabs.addresses")}</TabsTrigger>
          </TabsList>

          <TabsContent value="information" className="space-y-4">
            <ClientForm
              client={form}
              onClientChange={setForm}
              fieldErrors={fieldErrors}
            />
            <div className="flex justify-end">
              <Button type="button" onClick={handleSave} disabled={submitting}>
                {submitting ? tCommon("actions.saving") : tCommon("actions.save")}
              </Button>
            </div>
          </TabsContent>

          <TabsContent value="addresses">
            {selectedClient && (
              <AddressesTable ownerType="client" ownerId={selectedClient.client_id} />
            )}
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  );
}
