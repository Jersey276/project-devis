"use client";

import { useEffect, useState } from "react";
import {
  ResponsiveDialog,
  ResponsiveDialogBody,
  ResponsiveDialogContent,
  ResponsiveDialogFooter,
  ResponsiveDialogHeader,
  ResponsiveDialogTitle,
} from "@/components/custom/responsive-dialog";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";
import { PlusIcon } from "lucide-react";
import { toast } from "sonner";
import { createProject } from "@/lib/services/projects";
import { listClients } from "@/lib/services/clients";
import { clientLabel } from "@/lib/project-utils";
import type { BackendClient } from "@/types/backend";

export default function CreateProjectDialog({ onCreated }: { onCreated?: () => void }) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [clientId, setClientId] = useState("");
  const [clients, setClients] = useState<BackendClient[]>([]);
  const [nameError, setNameError] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (!open) return;
    listClients().then(({ ok, body }) => {
      if (ok && Array.isArray(body.clients)) setClients(body.clients);
    });
  }, [open]);

  function handleClose(v: boolean) {
    if (!v) {
      setName("");
      setClientId("");
      setNameError("");
    }
    setOpen(v);
  }

  async function handleSubmit() {
    if (!name.trim()) {
      setNameError("Le nom est requis.");
      return;
    }
    setNameError("");
    setBusy(true);
    const { ok, body } = await createProject({ name: name.trim(), clientId: clientId || undefined });
    setBusy(false);
    if (!ok) {
      toast.error(body?.message ?? "La création du projet a échoué.");
      return;
    }
    toast.success("Projet créé.");
    setOpen(false);
    setName("");
    setClientId("");
    onCreated?.();
  }

  return (
    <>
      <Button size="sm" onClick={() => setOpen(true)}>
        <PlusIcon />
        Nouveau projet
      </Button>
      <ResponsiveDialog open={open} onOpenChange={handleClose}>
        <ResponsiveDialogContent>
          <ResponsiveDialogHeader>
            <ResponsiveDialogTitle>Nouveau projet</ResponsiveDialogTitle>
          </ResponsiveDialogHeader>
          <ResponsiveDialogBody className="flex flex-col gap-4">
            <Field>
              <FieldLabel>Nom du projet</FieldLabel>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Mon projet"
              />
              {nameError && <FieldError>{nameError}</FieldError>}
            </Field>
            <Field>
              <FieldLabel>Client (optionnel)</FieldLabel>
              <Combobox value={clientId} onValueChange={setClientId}>
                <ComboboxInput placeholder="Rechercher un client…" />
                <ComboboxContent>
                  <ComboboxList>
                    <ComboboxEmpty>Aucun client.</ComboboxEmpty>
                    {clients.map((c) => (
                      <ComboboxItem key={c.client_id} value={c.client_id}>
                        {clientLabel(c)}
                      </ComboboxItem>
                    ))}
                  </ComboboxList>
                </ComboboxContent>
              </Combobox>
            </Field>
          </ResponsiveDialogBody>
          <ResponsiveDialogFooter>
            <Button variant="outline" onClick={() => handleClose(false)} disabled={busy}>
              Annuler
            </Button>
            <Button onClick={handleSubmit} disabled={busy}>
              Créer
            </Button>
          </ResponsiveDialogFooter>
        </ResponsiveDialogContent>
      </ResponsiveDialog>
    </>
  );
}
