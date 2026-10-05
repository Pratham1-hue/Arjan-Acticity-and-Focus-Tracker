import type { Kind } from "@/lib/types";
import { fmtDur } from "@/lib/format";
import { Empty, KIND_COLOR, KindDot, ShareBar } from "@/components/Card";

export interface TopItem {
  key: string;
  name: string;
  sub?: string;
  seconds: number;
  kind: Kind;
}

// Ranked list with share bars (used for apps and websites).
export function TopList({ items, empty }: { items: TopItem[]; empty: string }) {
  if (items.length === 0) return <Empty>{empty}</Empty>;
  const max = items[0].seconds;

  return (
    <ul className="space-y-3">
      {items.map((it) => (
        <li key={it.key}>
          <div className="flex items-baseline justify-between gap-3 text-sm">
            <span className="flex min-w-0 items-center gap-2">
              <KindDot kind={it.kind} />
              <span className="truncate text-zinc-200">{it.name}</span>
              {it.sub && (
                <span className="truncate text-xs text-zinc-500">{it.sub}</span>
              )}
            </span>
            <span className="shrink-0 tabular-nums text-zinc-400">
              {fmtDur(it.seconds)}
            </span>
          </div>
          <ShareBar pct={(it.seconds / max) * 100} color={KIND_COLOR[it.kind]} />
        </li>
      ))}
    </ul>
  );
}
