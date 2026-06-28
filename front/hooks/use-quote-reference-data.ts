"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { apiFetch } from "@/lib/api";
import { listClients, getMyClients, getMyClientAddresses } from "@/lib/services/clients";
import { listAddresses } from "@/lib/services/addresses";
import { listAvailableTaxesForUser } from "@/lib/services/taxes";
import type { BackendAddress, BackendClient, BackendTax } from "@/types/backend";
import type { FormItem } from "@/hooks/use-quote-lines";

type UseQuoteReferenceDataParams = {
  clientId: string;
  userAddressId: number | null;
  isCustomer: boolean;
  loading: boolean;
  items: FormItem[];
};

export function useQuoteReferenceData({
  clientId,
  userAddressId,
  isCustomer,
  loading,
  items,
}: UseQuoteReferenceDataParams) {
  const [clients, setClients] = useState<BackendClient[]>([]);
  const [userId, setUserId] = useState("");
  const [userAddresses, setUserAddresses] = useState<BackendAddress[]>([]);
  const [addresses, setAddresses] = useState<BackendAddress[]>([]);
  const [availableTaxes, setAvailableTaxes] = useState<BackendTax[]>([]);

  // Load clients (provider mode) or customer profile (customer mode)
  useEffect(() => {
    let cancelled = false;
    if (isCustomer) {
      getMyClients().then(({ ok, body }) => {
        if (cancelled) return;
        if (ok && Array.isArray(body.clients) && body.clients.length > 0) {
          setClients(body.clients as BackendClient[]);
        }
      });
    } else {
      listClients().then(({ ok, body }) => {
        if (cancelled) return;
        if (ok && Array.isArray(body.clients)) {
          setClients(body.clients as BackendClient[]);
        }
      });
    }
    return () => { cancelled = true; };
  }, [isCustomer]);

  // Load current user id + their addresses (provider only)
  useEffect(() => {
    if (isCustomer) return;
    let cancelled = false;
    (async () => {
      const meRes = await apiFetch("/api/users/me");
      if (cancelled || !meRes.ok || !meRes.body.success || !meRes.body.user) return;
      const meId = (meRes.body.user as { user_id: string }).user_id;
      setUserId(meId);
      const { ok, body } = await listAddresses({ type: "user", userId: meId });
      if (cancelled) return;
      setUserAddresses(ok && Array.isArray(body.addresses) ? (body.addresses as BackendAddress[]) : []);
    })();
    return () => { cancelled = true; };
  }, [isCustomer]);

  // Load addresses for selected client
  useEffect(() => {
    let cancelled = false;
    if (isCustomer) {
      getMyClientAddresses().then(({ ok, body }) => {
        if (cancelled) return;
        setAddresses(ok && Array.isArray(body.addresses) ? (body.addresses as BackendAddress[]) : []);
      });
    } else {
      if (!clientId) return;
      listAddresses({ type: "client", clientId }).then(({ ok, body }) => {
        if (cancelled) return;
        setAddresses(ok && Array.isArray(body.addresses) ? (body.addresses as BackendAddress[]) : []);
      });
    }
    return () => { cancelled = true; };
  }, [clientId, isCustomer]);

  const clientAddresses = useMemo(() => (clientId ? addresses : []), [clientId, addresses]);

  // Compute which orphaned tax IDs (superseded) the lines still reference,
  // so the taxes fetch can include them and keep their labels renderable.
  const currentTaxIds = useMemo(
    () => new Set(availableTaxes.filter((t) => !t.superseded_at).map((t) => t.id)),
    [availableTaxes],
  );
  const includeTaxIds = useMemo(() => {
    const ids = items
      .map((i) => i.taxId)
      .filter((id): id is number => id != null && !currentTaxIds.has(id));
    return [...new Set(ids)].sort((a, b) => a - b);
  }, [items, currentTaxIds]);
  const includeTaxIdsKey = includeTaxIds.join(",");

  // Load available taxes; gated on `loading` to avoid a double-fetch in edit mode.
  useEffect(() => {
    if (loading) return;
    let cancelled = false;
    listAvailableTaxesForUser(includeTaxIds, userAddressId ?? undefined).then(
      ({ ok, body }) => {
        if (cancelled) return;
        setAvailableTaxes(ok && Array.isArray(body.taxes) ? (body.taxes as BackendTax[]) : []);
      },
    );
    return () => { cancelled = true; };
    // includeTaxIds is reference-unstable; key via the joined string.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loading, includeTaxIdsKey, userAddressId]);

  const taxById = useMemo(
    () => new Map(availableTaxes.map((t) => [t.id, t])),
    [availableTaxes],
  );

  const defaultTaxId = useMemo(
    () => availableTaxes.find((t) => t.is_default)?.id ?? null,
    [availableTaxes],
  );

  const refreshClients = useCallback(async () => {
    if (isCustomer) {
      const { ok, body } = await getMyClients();
      if (ok && Array.isArray(body.clients)) setClients(body.clients as BackendClient[]);
    } else {
      const { ok, body } = await listClients();
      if (ok && Array.isArray(body.clients)) setClients(body.clients as BackendClient[]);
    }
  }, [isCustomer]);

  const refreshUserAddresses = useCallback(async () => {
    if (!userId) return;
    const { ok, body } = await listAddresses({ type: "user", userId });
    if (ok && Array.isArray(body.addresses)) setUserAddresses(body.addresses as BackendAddress[]);
  }, [userId]);

  const refreshClientAddresses = useCallback(async () => {
    if (isCustomer) {
      const { ok, body } = await getMyClientAddresses();
      if (ok && Array.isArray(body.addresses)) setAddresses(body.addresses as BackendAddress[]);
    } else {
      if (!clientId) return;
      const { ok, body } = await listAddresses({ type: "client", clientId });
      if (ok && Array.isArray(body.addresses)) setAddresses(body.addresses as BackendAddress[]);
    }
  }, [clientId, isCustomer]);

  return {
    clients,
    userId,
    userAddresses,
    addresses: clientAddresses,
    availableTaxes,
    taxById,
    defaultTaxId,
    refreshClients,
    refreshUserAddresses,
    refreshClientAddresses,
  };
}
