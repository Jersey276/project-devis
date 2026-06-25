"use client";

import { createContext, useCallback, useContext, useState } from "react";

export type UserMode = "provider" | "customer";

const MODE_COOKIE = "user-mode";


function writeModeCookie(mode: UserMode) {
  document.cookie = `${MODE_COOKIE}=${mode}; path=/; max-age=31536000; SameSite=Lax`;
}

type ModeContextValue = {
  mode: UserMode;
  setMode: (mode: UserMode) => void;
  isProvider: boolean;
  isCustomer: boolean;
};

const ModeContext = createContext<ModeContextValue | null>(null);

export function ModeProvider({
  initialMode,
  children,
}: {
  initialMode?: UserMode;
  children: React.ReactNode;
}) {
  const [mode, setModeState] = useState<UserMode>(initialMode ?? "provider");

  const setMode = useCallback((next: UserMode) => {
    setModeState((current) => {
      if (current === next) return current;
      writeModeCookie(next);
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
