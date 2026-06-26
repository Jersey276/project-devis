"use client";

import {
  PieChart,
  Pie,
  Cell,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";

export type PieChartDataPoint = {
  name: string;
  value: number;
  color: string;
};

type Props = {
  title: string;
  data: PieChartDataPoint[];
  innerRadius?: number;
  outerRadius?: number;
  height?: number;
  showLegend?: boolean;
  tooltipSuffix?: string;
};

export default function PieChartCard({
  title,
  data,
  innerRadius = 0,
  outerRadius = 70,
  height = 220,
  showLegend = false,
  tooltipSuffix = "",
}: Props) {
  return (
    <div className="rounded-lg border bg-card p-4">
      <p className="mb-3 text-sm font-medium">{title}</p>
      <ResponsiveContainer width="100%" height={height}>
        <PieChart>
          <Pie
            data={data}
            cx="50%"
            cy="50%"
            innerRadius={innerRadius}
            outerRadius={outerRadius}
            dataKey="value"
            label={({ name, value }) => `${name} (${value})`}
            labelLine={false}
          >
            {data.map((entry, i) => (
              <Cell key={i} fill={entry.color} />
            ))}
          </Pie>
          <Tooltip
            formatter={(v) => [`${v}${tooltipSuffix}`, "Nb"]}
          />
          {showLegend && <Legend />}
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
}
