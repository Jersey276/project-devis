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

// Factur-X: hybrid PDF/A-3 with the embedded EN 16931 XML. Issued invoices only.
export async function exportInvoiceFacturx(invoiceId: string): Promise<void> {
  await downloadBlob(
    `/api/export/invoices/${encodeURIComponent(invoiceId)}?facturx=1`,
    `facture-${invoiceId}.pdf`,
  );
}

export async function exportCreditNotePdf(creditNoteId: string): Promise<void> {
  await downloadBlob(
    `/api/export/credit-notes/${encodeURIComponent(creditNoteId)}`,
    `avoir-${creditNoteId}.pdf`,
  );
}

// Factur-X: hybrid PDF/A-3 with the embedded EN 16931 XML (type 381). Issued
// credit notes only.
export async function exportCreditNoteFacturx(creditNoteId: string): Promise<void> {
  await downloadBlob(
    `/api/export/credit-notes/${encodeURIComponent(creditNoteId)}?facturx=1`,
    `avoir-${creditNoteId}.pdf`,
  );
}
