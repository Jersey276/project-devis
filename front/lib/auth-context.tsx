"use client";

import { createContext, useContext } from "react";
import type { AuthContext } from "@/lib/access";

type AuthContextValue = {
  auth: AuthContext | null;
  ok: boolean;
};

const AuthCtx = createContext<AuthContextValue | null>(null);

export function AuthProvider({
  auth,
  ok,
  children,
}: {
  auth: AuthContext | null;
  ok: boolean;
  children: React.ReactNode;
}) {
  return <AuthCtx.Provider value={{ auth, ok }}>{children}</AuthCtx.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthCtx);
  if (!ctx) throw new Error("useAuth must be used within an AuthProvider");
  return ctx;
}
