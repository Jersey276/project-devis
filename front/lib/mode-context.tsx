"use client";

import { createContext, useCallback, useContext, useState } from "react";

export type UserMode = "provider" | "customer";

type ModeContextValue = {
  mode: UserMode;
  setMode: (mode: UserMode) => void;
  isProvider: boolean;
  isCustomer: boolean;
};

const ModeContext = createContext<ModeContextValue | null>(null);

// Customer mode is currently disabled at the UI level: the sidebar toggle is
// removed and the provider always reports "provider". The full API (setMode,
// isCustomer) stays in place so re-enabling later is just a matter of
// restoring the toggle and the cookie reconciliation.
export function ModeProvider({ children }: { children: React.ReactNode }) {
  const [mode, setModeState] = useState<UserMode>("provider");

  const setMode = useCallback((next: UserMode) => {
    setModeState((current) => (current === next ? current : next));
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
