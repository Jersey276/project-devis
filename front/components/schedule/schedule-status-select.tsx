"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
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

export const SCHEDULE_STATUSES: BackendScheduleStatus[] = [
  "DRAFT",
  "NEGOCIATE",
  "DENIED",
  "VALID",
];

type ScheduleStatusSelectProps = {
  scheduleId: string;
  value: BackendScheduleStatus;
  onUpdated?: (nextStatus: BackendScheduleStatus) => void | Promise<void>;
  onError?: (message: string) => void;
  className?: string;
  disabled?: boolean;
};

export default function ScheduleStatusSelect({
  scheduleId,
  value,
  onUpdated,
  onError,
  className,
  disabled,
}: ScheduleStatusSelectProps) {
  const t = useTranslations("schedule.statusSelect");
  const tStatus = useTranslations("status.schedule");
  const [optimisticStatus, setOptimisticStatus] = useState<BackendScheduleStatus | null>(null);
  const status = optimisticStatus ?? value;
  const [isUpdating, setIsUpdating] = useState(false);

  function confirmationMessage(nextStatus: BackendScheduleStatus): string | null {
    switch (nextStatus) {
      case "VALID":
        return t("confirmValid");
      case "DENIED":
        return t("confirmDenied");
      default:
        return null;
    }
  }

  async function handleChange(nextValue: string) {
    const nextStatus = nextValue as BackendScheduleStatus;
    if (nextStatus === status || isUpdating || disabled) return;

    const confirmation = confirmationMessage(nextStatus);
    if (confirmation && !window.confirm(confirmation)) {
      return;
    }

    setOptimisticStatus(nextStatus);
    setIsUpdating(true);

    try {
      const { ok, body } = await updateScheduleStatus(scheduleId, {
        status: nextStatus,
      } satisfies UpdateScheduleStatusPayload);
      if (!ok || !body.success) {
        throw new Error((body.message as string) ?? t("updateError"));
      }
      await onUpdated?.(nextStatus);
    } catch (error) {
      onError?.(error instanceof Error ? error.message : t("updateError"));
    } finally {
      setOptimisticStatus(null);
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
        <SelectValue placeholder={t("placeholder")} />
      </SelectTrigger>
      <SelectContent>
        {SCHEDULE_STATUSES.map((option) => (
          <SelectItem key={option} value={option}>
            {tStatus(option)}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
