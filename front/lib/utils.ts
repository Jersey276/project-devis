import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatEurosFromCents(cents: number): string {
  return (cents / 100).toFixed(2) + " €";
}

export function formatDateFR(value: string | null | undefined): string {
  if (!value) return "—";
  return new Date(value).toLocaleDateString("fr-FR", { timeZone: "UTC" });
}

export function formatTimestampDateFR(
  value: string | null | undefined,
): string {
  if (!value) return "—";
  return new Date(value).toLocaleDateString("fr-FR");
}

export function formatDateTimeFR(value: string | null | undefined): string {
  if (!value) return "—";
  return new Date(value).toLocaleString("fr-FR", {
    dateStyle: "short",
    timeStyle: "short",
  });
}
