"use client";

type Props = {
  label: string;
  value: string | number;
};

export default function MetricCard({ label, value }: Props) {
  return (
    <div className="rounded-lg border bg-card p-4 text-card-foreground shadow-sm">
      <p className="text-sm text-muted-foreground">{label}</p>
      <p className="mt-1 text-2xl font-bold">{value}</p>
    </div>
  );
}
