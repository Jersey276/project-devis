"use client";

import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";

export type LineSeriesConfig = {
  key: string;
  color: string;
  label?: string;
};

type Props = {
  title: string;
  data: Record<string, string | number>[];
  lines: LineSeriesConfig[];
  xAxisKey: string;
  height?: number;
  xTickFormatter?: (v: string) => string;
  yTickFormatter?: (v: number) => string;
  tooltipLabelFormatter?: (label: string) => string;
  tooltipFormatter?: (value: number, name: string) => [string | number, string];
  vertical?: boolean;
};

export default function LineChartCard({
  title,
  data,
  lines,
  xAxisKey,
  height = 240,
  xTickFormatter,
  yTickFormatter,
  tooltipLabelFormatter,
  tooltipFormatter,
  vertical = true,
}: Props) {
  return (
    <div className="rounded-lg border bg-card p-4">
      <p className="mb-3 text-sm font-medium">{title}</p>
      <ResponsiveContainer width="100%" height={height}>
        <LineChart data={data} margin={{ top: 4, right: 8, left: 0, bottom: 4 }}>
          <CartesianGrid strokeDasharray="3 3" vertical={vertical} />
          <XAxis
            dataKey={xAxisKey}
            tickFormatter={xTickFormatter}
            tick={{ fontSize: 11 }}
            tickLine={false}
            axisLine={false}
          />
          <YAxis
            allowDecimals={false}
            tick={{ fontSize: 11 }}
            tickLine={false}
            axisLine={false}
            width={40}
            tickFormatter={yTickFormatter}
          />
          <Tooltip
            formatter={tooltipFormatter ? (value, name) => tooltipFormatter(value as number, name as string) : undefined}
            labelFormatter={tooltipLabelFormatter ? (label) => tooltipLabelFormatter(label as string) : undefined}
          />
          {lines.map((s) => (
            <Line
              key={s.key}
              type="monotone"
              dataKey={s.key}
              stroke={s.color}
              strokeWidth={2}
              dot={false}
              activeDot={{ r: 4 }}
              name={s.label ?? s.key}
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
