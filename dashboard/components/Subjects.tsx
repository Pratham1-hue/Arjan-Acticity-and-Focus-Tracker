"use client";

import type { SubjectStat } from "@/lib/types";
import { fmtDur } from "@/lib/format";
import { Empty, ShareBar } from "@/components/Card";
import { PALETTE } from "@/components/CategoryDonut";

// "What you learned": productive time per subject, with the
// topics (or lecture/page titles) that made it up.
export function Subjects({
  subjects,
  aiOn,
}: {
  subjects: SubjectStat[];
  aiOn: boolean;
}) {
  if (subjects.length === 0) {
    return (
      <Empty>
        No study time recorded yet.
        {!aiOn &&
          " Turn on AI categorization in Settings to detect subjects and topics automatically."}
      </Empty>
    );
  }
  const max = subjects[0].seconds;

  return (
    <ul className="space-y-4">
      {subjects.map((s, i) => (
        <li key={s.name}>
          <div className="flex items-baseline justify-between gap-2">
            <span className="font-medium text-zinc-100">{s.name}</span>
            <span className="tabular-nums text-sm font-semibold text-emerald-400">
              {fmtDur(s.seconds)}
            </span>
          </div>
          <ShareBar
            pct={(s.seconds / max) * 100}
            color={PALETTE[i % PALETTE.length]}
          />
          {s.topics.length > 0 && (
            <div className="mt-2 flex flex-wrap gap-1.5">
              {s.topics.map((t) => (
                <span
                  key={t.name}
                  title={t.name}
                  className="max-w-full truncate rounded-md bg-zinc-800/80 px-2 py-0.5 text-xs text-zinc-300"
                >
                  {t.name}
                  <span className="text-zinc-500"> · {fmtDur(t.seconds)}</span>
                </span>
              ))}
            </div>
          )}
        </li>
      ))}
    </ul>
  );
}
