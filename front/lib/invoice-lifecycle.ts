import type { BackendInvoiceLifecycleStatus } from "@/types/backend";

type RealLifecycleStatus = Exclude<BackendInvoiceLifecycleStatus, "NONE">;

// Mirror of the backend transition table (backend/invoice/actions/lifecycle.go).
// Used only to populate the UI menu; the server remains authoritative.
const TRANSITIONS: Record<BackendInvoiceLifecycleStatus, RealLifecycleStatus[]> =
  {
    NONE: ["DEPOSITED"],
    DEPOSITED: ["RECEIVED", "REJECTED"],
    RECEIVED: ["APPROVED", "REJECTED"],
    APPROVED: ["COLLECTED", "REJECTED"],
    REJECTED: [],
    COLLECTED: [],
  };

export function allowedNextLifecycleStatuses(
  current: BackendInvoiceLifecycleStatus,
): RealLifecycleStatus[] {
  return TRANSITIONS[current];
}
