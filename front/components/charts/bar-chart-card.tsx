"use client";

import {
  BarChart,
  Bar,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";

export type BarChartDataPoint = {
  name: string;
  [key: string]: string | number;
};

type Props = {
  title: string;
  data: BarChartDataPoint[];
  dataKey: string;
  colorKey?: string;
  defaultColor?: string;
  height?: number;
  tickFormatter?: (v: number) => string;
  tooltipFormatter?: (v: number) => string;
  barName?: string;
  noDataMessage?: string;
};

export default function BarChartCard({
  title,
  data,
  dataKey,
  colorKey,
  defaultColor = "#3b82f6",
  height = 200,
  tickFormatter,
  tooltipFormatter,
  barName,
  noDataMessage,
}: Props) {
  const hasData = data.length > 0;

  return (
    <div className="rounded-lg border bg-card p-4">
      <p className="mb-3 text-sm font-medium">{title}</p>
      {hasData ? (
        <ResponsiveContainer width="100%" height={height}>
          <BarChart data={data} margin={{ top: 4, right: 4, left: -20, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="name" tick={{ fontSize: 11 }} />
            <YAxis
              allowDecimals={false}
              tick={{ fontSize: 11 }}
              tickFormatter={tickFormatter}
            />
            <Tooltip
              formatter={(v) =>
                tooltipFormatter
                  ? [tooltipFormatter(v as number), barName ?? dataKey]
                  : [v, barName ?? dataKey]
              }
            />
            <Bar dataKey={dataKey} fill={defaultColor} name={barName}>
              {colorKey &&
                data.map((entry, i) => (
                  <Cell key={i} fill={String(entry[colorKey] ?? defaultColor)} />
                ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      ) : (
        noDataMessage && (
          <p className="mt-4 text-sm text-muted-foreground">{noDataMessage}</p>
        )
      )}
    </div>
  );
}
