import { useMemo } from "react";
import type { BackendScheduleDetails, BackendScheduleLineSummary } from "@/types/backend";
import {
  type ScheduleMonthHeader,
  type ScheduleYearHeader,
  buildScheduleMonthHeaders,
  buildScheduleYearHeaders,
} from "./schedule-utils";

type ScheduleTableStructure = {
  monthHeaders: ScheduleMonthHeader[];
  yearHeaders: ScheduleYearHeader[];
  columnTotalsByMonth: Map<number, number>;
  sortedLines: BackendScheduleLineSummary[];
  editableLines: BackendScheduleLineSummary[];
  editableIndexByLineId: Map<string, number>;
};

export function useScheduleTableStructure(
  schedule: BackendScheduleDetails | null,
): ScheduleTableStructure {
  const monthHeaders = useMemo(
    () =>
      schedule
        ? buildScheduleMonthHeaders(
            schedule.start_month,
            schedule.duration_months,
          )
        : [],
    [schedule],
  );

  const yearHeaders = useMemo(
    () => buildScheduleYearHeaders(monthHeaders),
    [monthHeaders],
  );

  const columnTotalsByMonth = useMemo(() => {
    if (!schedule) return new Map<number, number>();
    return new Map(
      schedule.column_totals.map((column) => [
        column.month_index,
        column.amount_cents,
      ]),
    );
  }, [schedule]);

  const sortedLines = useMemo(
    () =>
      schedule
        ? [...schedule.lines].sort(
            (a, b) => (a.position ?? 0) - (b.position ?? 0),
          )
        : [],
    [schedule],
  );

  const editableLines = useMemo(
    () => sortedLines.filter((l) => l.data_kind !== "group"),
    [sortedLines],
  );

  const editableIndexByLineId = useMemo(
    () => new Map(editableLines.map((l, i) => [l.quote_line_id, i])),
    [editableLines],
  );

  return {
    monthHeaders,
    yearHeaders,
    columnTotalsByMonth,
    sortedLines,
    editableLines,
    editableIndexByLineId,
  };
}
