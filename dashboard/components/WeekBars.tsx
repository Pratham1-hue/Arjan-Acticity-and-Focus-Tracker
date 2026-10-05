"use client";

import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

import type { WeekDay } from "@/lib/types";
import { fmtDur } from "@/lib/format";

const tooltipStyle = {
  background: "#18181b",
  border: "1px solid #3f3f46",
  borderRadius: 8,
  color: "#e4e4e7",
  fontSize: 12,
};

// Last 7 days, stacked by kind. Click a day to open it.
export function WeekBars({
  days,
  onSelect,
}: {
  days: WeekDay[];
  onSelect?: (date: string) => void;
}) {
  return (
    <div className="h-64 w-full">
      <ResponsiveContainer width="100%" height="100%">
        <BarChart
          data={days}
          margin={{ top: 8, right: 8, left: -14, bottom: 0 }}
          onClick={(state) => {
            const idx = Number((state as { activeIndex?: unknown } | null)?.activeIndex);
            if (onSelect && Number.isInteger(idx) && days[idx]) onSelect(days[idx].date);
          }}
          style={{ cursor: onSelect ? "pointer" : undefined }}
        >
          <CartesianGrid strokeDasharray="3 3" stroke="#27272a" vertical={false} />
          <XAxis dataKey="label" stroke="#71717a" fontSize={12} tickLine={false} axisLine={false} />
          <YAxis
            stroke="#71717a"
            fontSize={12}
            tickLine={false}
            axisLine={false}
            tickFormatter={(v: number) => (v >= 3600 ? `${Math.round(v / 3600)}h` : `${Math.round(v / 60)}m`)}
          />
          <Tooltip
            cursor={{ fill: "rgba(255,255,255,0.04)" }}
            contentStyle={tooltipStyle}
            formatter={(value, name) => [fmtDur(Number(value ?? 0)), String(name)]}
          />
          <Bar dataKey="productive" name="Productive" stackId="a" fill="#34d399" />
          <Bar dataKey="distraction" name="Distraction" stackId="a" fill="#fb7185" />
          <Bar dataKey="other" name="Other" stackId="a" fill="#52525b" radius={[4, 4, 0, 0]} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
