import type { DayReport } from "@/lib/types";
import { fmtDur, pct } from "@/lib/format";
import { Empty } from "@/components/Card";

// The day's evaluation, in plain sentences.
// With AI on, subjects are real subjects ("Machine Learning"); without
// it they fall back to rules.json categories ("Web Dev"), so the
// wording changes accordingly.
export function buildReview(day: DayReport, isToday: boolean, aiOn: boolean): string[] {
  if (day.total === 0) return [];
  const when = isToday ? "today" : "that day";
  const out: string[] = [];

  out.push(
    `You were active for ${fmtDur(day.total)} ${when} — ${pct(day.productive, day.total)}% of it productive.`,
  );

  for (const s of day.bySubject.filter((x) => x.seconds >= 60).slice(0, 4)) {
    const topics = s.topics
      .slice(0, 2)
      .map((t) => t.name)
      .join(" and ");
    const mostly = topics ? ` — mostly ${topics}` : "";
    out.push(
      aiOn
        ? `You learned ${s.name} for ${fmtDur(s.seconds)}${mostly}.`
        : `You spent ${fmtDur(s.seconds)} on ${s.name}${mostly}.`,
    );
  }

  const apps = day.byApp.slice(0, 4).map((a) => `${a.name} ${fmtDur(a.seconds)}`);
  if (apps.length) out.push(`Apps you used most: ${apps.join(", ")}.`);

  const site = day.bySite[0];
  if (site) out.push(`Top website: ${site.site} (${fmtDur(site.seconds)}).`);

  if (day.distraction > 0) {
    const worst = day.activities.find((a) => a.kind === "distraction");
    out.push(
      `Distractions took ${fmtDur(day.distraction)}${worst ? `, mostly "${worst.label}"` : ""}.`,
    );
  }

  if (day.deepWork.length) {
    const longest = Math.max(...day.deepWork.map((b) => b.seconds));
    out.push(
      `${day.deepWork.length} deep work block${day.deepWork.length > 1 ? "s" : ""} — longest ${fmtDur(longest)}.`,
    );
  }
  return out;
}

export function Review({ day, isToday, aiOn }: { day: DayReport; isToday: boolean; aiOn: boolean }) {
  const lines = buildReview(day, isToday, aiOn);
  if (lines.length === 0) return <Empty>Nothing to review yet.</Empty>;
  return (
    <ul className="space-y-2 text-sm leading-relaxed text-zinc-300">
      {lines.map((l, i) => (
        <li key={i} className="flex gap-2">
          <span className="mt-2 h-1 w-1 shrink-0 rounded-full bg-indigo-400" />
          <span>{l}</span>
        </li>
      ))}
    </ul>
  );
}
