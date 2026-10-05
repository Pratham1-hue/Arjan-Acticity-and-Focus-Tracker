import type { DeepWorkBlock } from "@/lib/types";
import { fmtDur, timeFmt } from "@/lib/format";
import { Empty } from "@/components/Card";

export function DeepWorkList({ blocks }: { blocks: DeepWorkBlock[] }) {
  if (blocks.length === 0) {
    return (
      <Empty>
        No deep work blocks — an uninterrupted productive stretch of 45+
        minutes shows up here.
      </Empty>
    );
  }
  const max = Math.max(...blocks.map((b) => b.seconds));

  return (
    <ul className="grid gap-3 sm:grid-cols-2">
      {blocks.map((b, i) => (
        <li key={i} className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-4">
          <div className="flex items-baseline justify-between gap-2">
            <span className="font-mono text-sm text-zinc-200">
              {timeFmt(b.start)} – {timeFmt(b.end)}
            </span>
            <span className="text-sm font-semibold text-indigo-400">
              {fmtDur(b.seconds)}
            </span>
          </div>
          <div className="mt-1 flex items-center justify-between gap-2">
            <span className="truncate text-xs text-zinc-500" title={b.label}>
              {b.label}
            </span>
            <span className="shrink-0 rounded-full bg-indigo-500/10 px-2 py-0.5 text-xs font-medium text-indigo-300">
              {b.category}
            </span>
          </div>
          <div className="mt-2.5 h-1.5 overflow-hidden rounded-full bg-zinc-800">
            <div
              className="h-1.5 rounded-full bg-gradient-to-r from-indigo-500 to-indigo-400"
              style={{ width: `${Math.max(6, (b.seconds / max) * 100)}%` }}
            />
          </div>
        </li>
      ))}
    </ul>
  );
}
