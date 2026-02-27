export default function invoiceDetailPage({
  params,
}: {
  params: { uuid: string };
}) {
  return <div>Invoice Detail Page from {params.uuid}</div>;
}
