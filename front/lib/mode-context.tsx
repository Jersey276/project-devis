"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";

export type UserMode = "provider" | "customer";

const STORAGE_KEY = "app.user-mode";
const DEFAULT_MODE: UserMode = "provider";

type ModeContextValue = {
  mode: UserMode;
  setMode: (mode: UserMode) => void;
  isProvider: boolean;
  isCustomer: boolean;
};

const ModeContext = createContext<ModeContextValue | null>(null);

function parseMode(raw: string | null | undefined): UserMode {
  return raw === "customer" || raw === "provider" ? raw : DEFAULT_MODE;
}

function readCookieMode(): UserMode | null {
  if (typeof document === "undefined") return null;
  const match = document.cookie.match(
    new RegExp(`(?:^|; )${STORAGE_KEY}=([^;]*)`),
  );
  return match ? parseMode(decodeURIComponent(match[1])) : null;
}

function writeCookieMode(mode: UserMode) {
  if (typeof document === "undefined") return;
  // Persist for ~1 year. Path=/ so every route sees the same value.
  document.cookie = `${STORAGE_KEY}=${mode}; path=/; max-age=31536000; samesite=lax`;
}

export function ModeProvider({
  initialMode,
  children,
}: {
  initialMode?: UserMode;
  children: React.ReactNode;
}) {
  const [mode, setModeState] = useState<UserMode>(initialMode ?? DEFAULT_MODE);

  // Reconcile with the cookie on mount in case it changed between SSR and
  // hydration (e.g. another tab updated it). Server-supplied initialMode is
  // the source of truth for the first paint to avoid hydration mismatches.
  useEffect(() => {
    const stored = readCookieMode();
    if (stored && stored !== mode) setModeState(stored);
    // Intentionally run only once on mount.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const setMode = useCallback((next: UserMode) => {
    setModeState((current) => {
      if (current === next) return current;
      writeCookieMode(next);
      return next;
    });
  }, []);

  return (
    <ModeContext.Provider
      value={{
        mode,
        setMode,
        isProvider: mode === "provider",
        isCustomer: mode === "customer",
      }}
    >
      {children}
    </ModeContext.Provider>
  );
}

export function useMode(): ModeContextValue {
  const ctx = useContext(ModeContext);
  if (!ctx) throw new Error("useMode must be used within a ModeProvider");
  return ctx;
}
