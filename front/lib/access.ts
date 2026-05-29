export type AuthContext = {
  user_id: string;
  email: string;
  role: "free_user" | "super_admin" | string;
  account_status: "active" | "suspended" | string;
  subscription_tier: "free" | string;
};

export function isSuperAdmin(auth: AuthContext | null): boolean {
  return auth?.role === "super_admin" && auth?.account_status === "active";
}
