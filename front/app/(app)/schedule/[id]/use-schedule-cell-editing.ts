import { useState } from "react";
import type { BackendScheduleDetails, BackendScheduleLineSummary } from "@/types/backend";
import { updateScheduleCell } from "@/lib/services/schedules";
import {
  draftKey,
  eurosStringToCents,
  applyOptimisticCellUpdate,
  parseAmountEur,
} from "./schedule-utils";

type Params = {
  schedule: BackendScheduleDetails | null;
  setSchedule: React.Dispatch<React.SetStateAction<BackendScheduleDetails | null>>;
  editableLines: BackendScheduleLineSummary[];
  editableIndexByLineId: Map<string, number>;
  isReadOnly: boolean;
  setError: (msg: string | null) => void;
};

type CellEditing = {
  cellDrafts: Record<string, string>;
  setCellDrafts: React.Dispatch<React.SetStateAction<Record<string, string>>>;
  savedCellDrafts: Record<string, string>;
  setSavedCellDrafts: React.Dispatch<React.SetStateAction<Record<string, string>>>;
  cellErrors: Record<string, string>;
  setCellErrors: React.Dispatch<React.SetStateAction<Record<string, string>>>;
  onCellDraftChange: (quoteLineId: string, monthIndex: number, value: string) => void;
  focusCellInput: (lineIndex: number, monthIndex: number) => void;
  handleCellKeyDown: (e: React.KeyboardEvent<HTMLInputElement>, lineIndex: number, monthIndex: number) => void;
  saveCell: (quoteLineId: string, monthIndex: number) => Promise<void>;
};

export function useScheduleCellEditing({
  schedule,
  setSchedule,
  editableLines,
  isReadOnly,
  setError,
}: Params): CellEditing {
  const [cellDrafts, setCellDrafts] = useState<Record<string, string>>({});
  const [savedCellDrafts, setSavedCellDrafts] = useState<Record<string, string>>({});
  const [cellErrors, setCellErrors] = useState<Record<string, string>>({});

  function onCellDraftChange(quoteLineId: string, monthIndex: number, value: string) {
    const key = draftKey(quoteLineId, monthIndex);
    setCellDrafts((prev) => ({ ...prev, [key]: value }));
    setCellErrors((prev) => {
      if (!prev[key]) return prev;
      const next = { ...prev };
      delete next[key];
      return next;
    });
  }

  function focusCellInput(lineIndex: number, monthIndex: number) {
    const line = editableLines[lineIndex];
    if (!line) return;
    const input = document.querySelector<HTMLInputElement>(
      `input[name='cell-${line.quote_line_id}-m${monthIndex}']`,
    );
    input?.focus();
    input?.select();
  }

  function handleCellKeyDown(
    e: React.KeyboardEvent<HTMLInputElement>,
    lineIndex: number,
    monthIndex: number,
  ) {
    if (!schedule || isReadOnly) return;

    const maxLineIndex = editableLines.length - 1;
    const maxMonthIndex = schedule.duration_months;

    if (e.key === "Enter") {
      e.preventDefault();
      e.currentTarget.blur();
      if (lineIndex < maxLineIndex) {
        setTimeout(() => { focusCellInput(lineIndex + 1, monthIndex); }, 0);
      }
      return;
    }

    if (e.key === "Tab") {
      e.preventDefault();
      if (e.shiftKey) {
        if (monthIndex > 1) { focusCellInput(lineIndex, monthIndex - 1); return; }
        if (lineIndex > 0) { focusCellInput(lineIndex - 1, maxMonthIndex); }
        return;
      }
      if (monthIndex < maxMonthIndex) { focusCellInput(lineIndex, monthIndex + 1); return; }
      if (lineIndex < maxLineIndex) { focusCellInput(lineIndex + 1, 1); }
      return;
    }

    if (e.key === "ArrowRight") {
      e.preventDefault();
      if (monthIndex < maxMonthIndex) { focusCellInput(lineIndex, monthIndex + 1); }
      return;
    }

    if (e.key === "ArrowLeft") {
      e.preventDefault();
      if (monthIndex > 1) { focusCellInput(lineIndex, monthIndex - 1); }
      return;
    }

    if (e.key === "ArrowDown") {
      e.preventDefault();
      if (lineIndex < maxLineIndex) { focusCellInput(lineIndex + 1, monthIndex); }
      return;
    }

    if (e.key === "ArrowUp") {
      e.preventDefault();
      if (lineIndex > 0) { focusCellInput(lineIndex - 1, monthIndex); }
    }
  }

  async function saveCell(quoteLineId: string, monthIndex: number) {
    if (!schedule || isReadOnly) return;

    const key = draftKey(quoteLineId, monthIndex);
    const rawAmount = cellDrafts[key] ?? "";

    const parsedAmount = parseAmountEur(rawAmount);
    if (!parsedAmount.ok) {
      if (parsedAmount.reason === "empty") {
        setCellDrafts((prev) => ({ ...prev, [key]: savedCellDrafts[key] ?? "" }));
        setCellErrors((prev) => {
          if (!prev[key]) return prev;
          const next = { ...prev };
          delete next[key];
          return next;
        });
        setError(null);
        return;
      }
      setCellErrors((prev) => ({ ...prev, [key]: parsedAmount.message ?? "Montant invalide." }));
      setError(null);
      return;
    }

    const savedParsed = parseAmountEur(savedCellDrafts[key] ?? "");
    if (savedParsed.ok && savedParsed.normalizedValue === parsedAmount.normalizedValue) {
      setCellDrafts((prev) => ({ ...prev, [key]: parsedAmount.normalizedValue }));
      setCellErrors((prev) => {
        if (!prev[key]) return prev;
        const next = { ...prev };
        delete next[key];
        return next;
      });
      setError(null);
      return;
    }

    const { ok, body } = await updateScheduleCell(schedule.schedule_id, {
      quoteLineId,
      monthIndex,
      amountEur: parsedAmount.normalizedValue,
    });
    if (!ok || !body.success) {
      const message = (body.message as string) ?? "Mise à jour impossible.";
      setCellErrors((prev) => ({ ...prev, [key]: message }));
      setError(null);
      return;
    }

    const previousAmountCents = savedParsed.ok ? eurosStringToCents(savedParsed.normalizedValue) : 0;
    const amountCents = eurosStringToCents(parsedAmount.normalizedValue);

    setSchedule((prev) =>
      prev ? applyOptimisticCellUpdate(prev, quoteLineId, monthIndex, amountCents, previousAmountCents) : prev,
    );
    setCellDrafts((prev) => ({ ...prev, [key]: parsedAmount.normalizedValue }));
    setSavedCellDrafts((prev) => ({ ...prev, [key]: parsedAmount.normalizedValue }));
    setCellErrors((prev) => {
      if (!prev[key]) return prev;
      const next = { ...prev };
      delete next[key];
      return next;
    });
    setError(null);
  }

  return {
    cellDrafts,
    setCellDrafts,
    savedCellDrafts,
    setSavedCellDrafts,
    cellErrors,
    setCellErrors,
    onCellDraftChange,
    focusCellInput,
    handleCellKeyDown,
    saveCell,
  };
}
