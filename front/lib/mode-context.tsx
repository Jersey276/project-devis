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

function readStoredMode(): UserMode {
  if (typeof window === "undefined") return DEFAULT_MODE;
  const raw = window.localStorage.getItem(STORAGE_KEY);
  return raw === "customer" || raw === "provider" ? raw : DEFAULT_MODE;
}

export function ModeProvider({ children }: { children: React.ReactNode }) {
  const [mode, setModeState] = useState<UserMode>(readStoredMode);

  // Sync across tabs.
  useEffect(() => {
    function onStorage(event: StorageEvent) {
      if (event.key !== STORAGE_KEY) return;
      const next =
        event.newValue === "customer" || event.newValue === "provider"
          ? event.newValue
          : DEFAULT_MODE;
      setModeState(next);
    }
    window.addEventListener("storage", onStorage);
    return () => window.removeEventListener("storage", onStorage);
  }, []);

  const setMode = useCallback((next: UserMode) => {
    setModeState((current) => {
      if (current === next) return current;
      if (typeof window !== "undefined") {
        window.localStorage.setItem(STORAGE_KEY, next);
      }
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
