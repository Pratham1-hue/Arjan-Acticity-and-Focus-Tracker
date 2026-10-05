"use client";

import {
  Cell,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
} from "recharts";

// Shared category palette, also used for the legend.
export const PALETTE = [
  "#818cf8", // indigo
  "#34d399", // emerald
  "#fbbf24", // amber
  "#f472b6", // pink
  "#38bdf8", // sky
  "#a78bfa", // violet
  "#fb7185", // rose
  "#94a3b8", // slate
];

export interface PieDatum {
  name: string;
  seconds: number;
}

const tooltipStyle = {
  background: "#18181b",
  border: "1px solid #3f3f46",
  borderRadius: 8,
  color: "#e4e4e7",
  fontSize: 12,
};

export function CategoryDonut({ data }: { data: PieDatum[] }) {
  const chartData = data.map((d) => ({
    name: d.name,
    value: Math.round(d.seconds / 60),
  }));

  if (chartData.length === 0) {
    return (
      <div className="flex h-64 items-center justify-center text-sm text-zinc-500">
        No activity recorded yet today
      </div>
    );
  }

  return (
    <div className="h-64 w-full">
      <ResponsiveContainer width="100%" height="100%">
        <PieChart>
          <Pie
            data={chartData}
            dataKey="value"
            nameKey="name"
            innerRadius="58%"
            outerRadius="88%"
            paddingAngle={2}
            strokeWidth={0}
          >
            {chartData.map((_, i) => (
              <Cell key={i} fill={PALETTE[i % PALETTE.length]} />
            ))}
          </Pie>
          <Tooltip
            formatter={(value, name) => [
              `${Number(value ?? 0)} min`,
              String(name),
            ]}
            contentStyle={tooltipStyle}
          />
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
}
