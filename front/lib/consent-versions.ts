export const CONSENT_VERSIONS = {
  cgv: "2025-07",
  privacy_policy: "2025-07",
} as const;

export type ConsentType = keyof typeof CONSENT_VERSIONS;
