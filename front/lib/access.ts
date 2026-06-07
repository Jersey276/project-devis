export type AuthContext = {
  user_id: string;
  email: string;
  role: "free_user" | "super_admin" | string;
  account_status: "active" | "suspended" | string;
  subscription_tier: "free" | "pro" | "enterprise" | string;
};

export function isSuperAdmin(auth: AuthContext | null): boolean {
  return auth?.role === "super_admin" && auth?.account_status === "active";
}

export function isEnterprise(auth: AuthContext | null): boolean {
  return auth?.subscription_tier === "enterprise";
}

export function isPro(auth: AuthContext | null): boolean {
  return auth?.subscription_tier === "pro" || isEnterprise(auth);
}

export function canUsePaidFeatures(auth: AuthContext | null): boolean {
  return isPro(auth) || isEnterprise(auth);
}
