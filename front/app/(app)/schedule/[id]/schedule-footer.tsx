"use client";

import { TableCell, TableFooter, TableRow } from "@/components/ui/table";
import { formatEurosFromCents } from "@/lib/utils";
import type { BackendScheduleDetails } from "@/types/backend";
import type { ScheduleMonthHeader } from "./schedule-utils";

type Props = {
  schedule: BackendScheduleDetails;
  monthHeaders: ScheduleMonthHeader[];
  columnTotalsByMonth: Map<number, number>;
};

export default function ScheduleFooter({
  schedule,
  monthHeaders,
  columnTotalsByMonth,
}: Props) {
  return (
    <TableFooter>
      <TableRow>
        <TableCell className="sticky left-0 z-20 min-w-44 bg-muted/60 font-semibold">
          Totaux mensuels
        </TableCell>
        <TableCell
          data-testid="footer-total-quote"
          className="sticky left-44 z-20 min-w-32 bg-muted/60 text-center font-semibold"
        >
          {formatEurosFromCents(schedule.quote_total_cents)}
        </TableCell>
        <TableCell
          data-testid="footer-total-remaining"
          className="sticky left-76 z-20 min-w-32 bg-muted/60 text-center font-semibold"
        >
          {formatEurosFromCents(
            schedule.quote_total_cents - schedule.planned_total_cents,
          )}
        </TableCell>
        {monthHeaders.map((_, index) => {
          const monthIndex = index + 1;
          return (
            <TableCell
              key={`footer-month-total-${monthIndex}`}
              data-testid={`footer-month-total-${monthIndex}`}
              className="text-center font-semibold"
            >
              {formatEurosFromCents(
                columnTotalsByMonth.get(monthIndex) ?? 0,
              )}
            </TableCell>
          );
        })}
      </TableRow>
    </TableFooter>
  );
}
