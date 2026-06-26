"use client";

import type { BackendScheduleDetails } from "@/types/backend";
import ScheduleStatusSelect from "@/components/schedule/schedule-status-select";

type Props = {
  schedule: BackendScheduleDetails;
  isCustomer: boolean;
  onUpdated: () => void;
  onError: (msg: string) => void;
};

export default function ScheduleMetadata({
  schedule,
  isCustomer,
  onUpdated,
  onError,
}: Props) {
  return (
    <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-4">
      <p>
        <strong>Nom:</strong> {schedule.name}
      </p>
      <p>
        <strong>Statut:</strong>{" "}
        {isCustomer ? (
          <span>{schedule.status}</span>
        ) : (
          <ScheduleStatusSelect
            scheduleId={schedule.schedule_id}
            value={schedule.status}
            className="w-44"
            onUpdated={onUpdated}
            onError={onError}
          />
        )}
      </p>
      <p>
        <strong>Début:</strong> {schedule.start_month}
      </p>
      <p>
        <strong>Durée:</strong> {schedule.duration_months} mois
      </p>
    </div>
  );
}
