"use client";

import { cn, formatEurosFromCents } from "@/lib/utils";
import {
  TableCell,
  TableRow,
} from "@/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { AlertCircle } from "lucide-react";
import { type BackendScheduleLineSummary, scheduleBalanceState } from "@/types/backend";
import { balanceStateClasses, draftKey, type ScheduleMonthHeader } from "./schedule-utils";

type Props = {
  line: BackendScheduleLineSummary;
  monthHeaders: ScheduleMonthHeader[];
  editableIndexByLineId: Map<string, number>;
  cellDrafts: Record<string, string>;
  cellErrors: Record<string, string>;
  isReadOnly: boolean;
  onCellDraftChange: (quoteLineId: string, monthIndex: number, value: string) => void;
  onCellBlur: (quoteLineId: string, monthIndex: number) => void;
  onCellKeyDown: (e: React.KeyboardEvent<HTMLInputElement>, lineIndex: number, monthIndex: number) => void;
};

export default function ScheduleLineRow({
  line,
  monthHeaders,
  editableIndexByLineId,
  cellDrafts,
  cellErrors,
  isReadOnly,
  onCellDraftChange,
  onCellBlur,
  onCellKeyDown,
}: Props) {
  if (line.data_kind === "group") {
    return (
      <TableRow
        key={line.quote_line_id}
        className="bg-muted/20 hover:bg-muted/20"
      >
        <TableCell
          colSpan={3 + monthHeaders.length}
          className="sticky left-0 z-20 bg-muted/20 text-xs font-semibold uppercase tracking-wide text-muted-foreground"
        >
          {line.name || "(Sans nom)"}
        </TableCell>
      </TableRow>
    );
  }

  const editableIndex = editableIndexByLineId.get(line.quote_line_id) ?? -1;
  const isChild = Boolean(line.parent_line_id);
  const displayName = line.name ?? line.quote_line_id;
  const state = scheduleBalanceState(line.planned_cents, line.expected_cents);
  const stateClasses = balanceStateClasses(state);

  return (
    <TableRow key={line.quote_line_id} className={stateClasses.rowClass}>
      <TableCell
        className={cn(
          "sticky left-0 z-20 min-w-44",
          stateClasses.stickyCellClass,
          isChild && "pl-8",
        )}
      >
        {displayName}
      </TableCell>
      <TableCell
        data-testid={`line-total-${line.quote_line_id}`}
        className={cn(
          "sticky left-44 z-20 min-w-32 text-center",
          stateClasses.stickyCellClass,
        )}
      >
        {formatEurosFromCents(line.expected_cents)}
      </TableCell>
      <TableCell
        data-testid={`line-remaining-${line.quote_line_id}`}
        className={cn(
          "sticky left-76 z-20 min-w-32 text-center",
          stateClasses.stickyCellClass,
        )}
      >
        {formatEurosFromCents(line.expected_cents - line.planned_cents)}
      </TableCell>
      {monthHeaders.map((_, index) => {
        const monthIndex = index + 1;
        const key = draftKey(line.quote_line_id, monthIndex);
        const cellError = cellErrors[key];
        return (
          <TableCell key={`${line.quote_line_id}-month-${monthIndex}`}>
            <div className="relative inline-flex items-center">
              <input
                name={`cell-${line.quote_line_id}-m${monthIndex}`}
                className={cn(
                  "h-9 w-24 rounded-md border px-2",
                  cellError
                    ? "border-destructive pr-8 focus-visible:ring-destructive"
                    : null,
                )}
                aria-invalid={Boolean(cellError)}
                disabled={isReadOnly}
                value={cellDrafts[key] ?? ""}
                onChange={(e) =>
                  onCellDraftChange(line.quote_line_id, monthIndex, e.target.value)
                }
                onBlur={() => onCellBlur(line.quote_line_id, monthIndex)}
                onKeyDown={(e) => onCellKeyDown(e, editableIndex, monthIndex)}
              />
              {cellError ? (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <button
                      type="button"
                      data-testid={`cell-error-${line.quote_line_id}-m${monthIndex}`}
                      className="text-destructive absolute right-2 inline-flex h-4 w-4 items-center justify-center"
                      aria-label={cellError}
                      title={cellError}
                    >
                      <AlertCircle className="h-4 w-4" />
                    </button>
                  </TooltipTrigger>
                  <TooltipContent side="top" sideOffset={6}>
                    {cellError}
                  </TooltipContent>
                </Tooltip>
              ) : null}
            </div>
          </TableCell>
        );
      })}
    </TableRow>
  );
}
