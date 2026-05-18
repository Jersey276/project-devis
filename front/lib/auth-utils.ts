export const NEXT_PARAM = "next";

// Reject anything that isn't a same-origin path to prevent open-redirect via ?next=.
export function safeNextPath(value: string | null): string {
  if (!value || !value.startsWith("/quote") || value.startsWith("//"))
    return "/quote";
  return value;
}
