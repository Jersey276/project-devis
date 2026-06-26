"use client";

import { TableHead, TableHeader, TableRow } from "@/components/ui/table";
import type { ScheduleMonthHeader, ScheduleYearHeader } from "./schedule-utils";

type Props = {
  yearHeaders: ScheduleYearHeader[];
  monthHeaders: ScheduleMonthHeader[];
};

export default function ScheduleTableHeader({ yearHeaders, monthHeaders }: Props) {
  return (
    <TableHeader className="bg-muted/40">
      <TableRow className="text-left">
        <TableHead
          rowSpan={2}
          className="sticky left-0 z-30 min-w-44 bg-muted/40"
        >
          Ligne devis
        </TableHead>
        <TableHead
          rowSpan={2}
          className="sticky left-44 z-30 min-w-32 bg-muted/40 text-center"
        >
          Total
        </TableHead>
        <TableHead
          rowSpan={2}
          className="sticky left-76 z-30 min-w-32 bg-muted/40 text-center"
        >
          Restant
        </TableHead>
        {yearHeaders.map((header) => (
          <TableHead
            key={`head-year-${header.year}`}
            colSpan={header.span}
            className="text-center"
          >
            {header.year}
          </TableHead>
        ))}
      </TableRow>
      <TableRow className="text-left">
        {monthHeaders.map((header, index) => (
          <TableHead
            key={`head-month-${index + 1}`}
            className="text-center uppercase"
          >
            {header.label}
          </TableHead>
        ))}
      </TableRow>
    </TableHeader>
  );
}
