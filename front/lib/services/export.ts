import { downloadBlob } from "@/lib/download";

export async function exportQuotePdf(quoteId: string): Promise<void> {
  await downloadBlob(
    `/api/export/quotes/${encodeURIComponent(quoteId)}`,
    `devis-${quoteId}.pdf`,
  );
}
