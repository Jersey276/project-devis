import { useCallback, useState } from "react";

export function useReloadKey() {
  const [key, setKey] = useState(0);
  const reload = useCallback(() => setKey((k) => k + 1), []);
  return { key, reload };
}
