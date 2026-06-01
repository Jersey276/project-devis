"use client";

import { useEffect, useState } from "react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  updateScheduleStatus,
  type UpdateScheduleStatusPayload,
} from "@/lib/services/schedules";
import type { BackendScheduleStatus } from "@/types/backend";

const STATUS_LABELS: Record<BackendScheduleStatus, string> = {
  DRAFT: "Brouillon",
  NEGOCIATE: "En négociation",
  DENIED: "Refusé",
  VALID: "Validé",
};

type ScheduleStatusSelectProps = {
  scheduleId: string;
  value: BackendScheduleStatus;
  onUpdated?: (nextStatus: BackendScheduleStatus) => void | Promise<void>;
  onError?: (message: string) => void;
  className?: string;
  disabled?: boolean;
};

function confirmationMessage(status: BackendScheduleStatus): string | null {
  switch (status) {
    case "VALID":
      return "Confirmer la validation de cet échéancier ? Cette action est definitive.";
    case "DENIED":
      return "Confirmer le refus de cet échéancier ?";
    default:
      return null;
  }
}

export default function ScheduleStatusSelect({
  scheduleId,
  value,
  onUpdated,
  onError,
  className,
  disabled,
}: ScheduleStatusSelectProps) {
  const [status, setStatus] = useState<BackendScheduleStatus>(value);
  const [isUpdating, setIsUpdating] = useState(false);

  useEffect(() => {
    setStatus(value);
  }, [value]);

  async function handleChange(nextValue: string) {
    const nextStatus = nextValue as BackendScheduleStatus;
    if (nextStatus === status || isUpdating || disabled) return;

    const confirmation = confirmationMessage(nextStatus);
    if (confirmation && !window.confirm(confirmation)) {
      setStatus(value);
      return;
    }

    const previousStatus = status;
    setStatus(nextStatus);
    setIsUpdating(true);

    try {
      const { ok, body } = await updateScheduleStatus(scheduleId, {
        status: nextStatus,
      } satisfies UpdateScheduleStatusPayload);
      if (!ok || !body.success) {
        throw new Error((body.message as string) ?? "Mise à jour impossible.");
      }
      await onUpdated?.(nextStatus);
    } catch (error) {
      setStatus(previousStatus);
      onError?.(
        error instanceof Error ? error.message : "Mise à jour impossible.",
      );
    } finally {
      setIsUpdating(false);
    }
  }

  return (
    <Select
      value={status}
      onValueChange={handleChange}
      disabled={disabled || isUpdating}
    >
      <SelectTrigger className={className} size="sm">
        <SelectValue placeholder="Statut" />
      </SelectTrigger>
      <SelectContent>
        {(Object.keys(STATUS_LABELS) as BackendScheduleStatus[]).map(
          (option) => (
            <SelectItem key={option} value={option}>
              {STATUS_LABELS[option]}
            </SelectItem>
          ),
        )}
      </SelectContent>
    </Select>
  );
}
