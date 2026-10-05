import type { ActivityStat } from "@/lib/types";
import { fmtDur } from "@/lib/format";
import { Empty, KindDot } from "@/components/Card";

// "What you did": the specific pages, lectures and windows,
// longest first.
export function Activities({ items }: { items: ActivityStat[] }) {
  if (items.length === 0) return <Empty>Nothing recorded for this day.</Empty>;

  return (
    <div className="divide-y divide-zinc-800/80">
      {items.map((a, i) => (
        <div key={i} className="flex items-center gap-3 py-2.5">
          <KindDot kind={a.kind} />
          <div className="min-w-0 flex-1">
            <div className="truncate text-sm text-zinc-200" title={a.label}>
              {a.label}
            </div>
            <div className="truncate text-xs text-zinc-500">
              {a.app}
              {a.site && a.site !== a.label ? ` · ${a.site}` : ""}
              {a.topic ? ` · ${a.topic}` : ""}
            </div>
          </div>
          <span className="hidden max-w-[12rem] shrink-0 truncate rounded-full bg-zinc-800 px-2 py-0.5 text-xs text-zinc-300 sm:inline">
            {a.kind === "productive" ? a.subject || a.category : a.category}
          </span>
          <span className="w-16 shrink-0 text-right text-sm tabular-nums text-zinc-400">
            {fmtDur(a.seconds)}
          </span>
        </div>
      ))}
    </div>
  );
}
