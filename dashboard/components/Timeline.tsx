import type { TimelineBlock } from "@/lib/types";
import { fmtDur, timeFmt } from "@/lib/format";
import { Empty, KIND_COLOR } from "@/components/Card";

// A horizontal strip of the day: each block is one activity,
// colored by kind. Hover for details.
export function Timeline({
  blocks,
  firstTs,
  lastTs,
}: {
  blocks: TimelineBlock[];
  firstTs: number;
  lastTs: number;
}) {
  if (blocks.length === 0) return <Empty>No activity on this day.</Empty>;

  // Snap the range to local hour boundaries.
  const s = new Date(firstTs * 1000);
  s.setMinutes(0, 0, 0);
  const e = new Date(lastTs * 1000);
  if (e.getMinutes() || e.getSeconds()) {
    e.setMinutes(0, 0, 0);
    e.setHours(e.getHours() + 1);
  }
  const start = s.getTime() / 1000;
  const end = Math.max(e.getTime() / 1000, start + 3600);
  const span = end - start;

  const ticks: number[] = [];
  for (const d = new Date(s); d.getTime() / 1000 <= end; d.setHours(d.getHours() + 1)) {
    ticks.push(d.getTime() / 1000);
  }
  const every = Math.max(1, Math.ceil(ticks.length / 12));
  const pos = (t: number) => ((t - start) / span) * 100;

  return (
    <div className="px-1">
      <div className="relative h-10 overflow-hidden rounded-lg bg-zinc-800/60">
        {blocks.map((b, i) => (
          <div
            key={i}
            className="absolute top-0 h-full hover:brightness-125"
            style={{
              left: `${pos(b.start)}%`,
              width: `${Math.max(pos(b.end) - pos(b.start), 0.2)}%`,
              backgroundColor: KIND_COLOR[b.kind],
            }}
            title={`${timeFmt(b.start)}–${timeFmt(b.end)} · ${b.app} · ${b.label} (${fmtDur(b.end - b.start)})`}
          />
        ))}
      </div>
      <div className="relative mt-1.5 h-4 text-[10px] text-zinc-500">
        {ticks.map(
          (t, i) =>
            i % every === 0 && (
              <span
                key={t}
                className="absolute -translate-x-1/2 tabular-nums"
                style={{ left: `${pos(t)}%` }}
              >
                {timeFmt(t)}
              </span>
            ),
        )}
      </div>
    </div>
  );
}
