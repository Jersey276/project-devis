import { downloadBlob } from "@/lib/download";

export async function exportQuotePdf(quoteId: string): Promise<void> {
  await downloadBlob(
    `/api/export/quotes/${encodeURIComponent(quoteId)}`,
    `devis-${quoteId}.pdf`,
  );
}

export async function exportSchedulePdf(scheduleId: string): Promise<void> {
  await downloadBlob(
    `/api/export/schedules/${encodeURIComponent(scheduleId)}`,
    `echeancier-${scheduleId}.pdf`,
  );
}

export async function exportInvoicePdf(invoiceId: string): Promise<void> {
  await downloadBlob(
    `/api/export/invoices/${encodeURIComponent(invoiceId)}`,
    `facture-${invoiceId}.pdf`,
  );
}

export async function exportCreditNotePdf(creditNoteId: string): Promise<void> {
  await downloadBlob(
    `/api/export/credit-notes/${encodeURIComponent(creditNoteId)}`,
    `avoir-${creditNoteId}.pdf`,
  );
}
