"use client";

import { useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";
import { type Country } from "@/components/address/address-form";

export function useCountries(reloadKey = 0, skip = false): Country[] {
  const [countries, setCountries] = useState<Country[]>([]);

  useEffect(() => {
    if (skip) return;
    let cancelled = false;
    apiFetch("/api/users/countries").then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && Array.isArray(body.countries)) {
        setCountries(body.countries as Country[]);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [reloadKey, skip]);

  return countries;
}
