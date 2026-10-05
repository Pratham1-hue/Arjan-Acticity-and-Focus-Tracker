/** "2h 05m", "45m", "12s" */
export function fmtDur(seconds: number): string {
  const totalMin = Math.round(seconds / 60);
  if (totalMin < 1) return `${Math.round(seconds)}s`;
  const h = Math.floor(totalMin / 60);
  const m = totalMin % 60;
  if (h === 0) return `${m}m`;
  return `${h}h ${String(m).padStart(2, "0")}m`;
}

/** "14:05" (24h, local time) */
export function timeFmt(ts: number): string {
  return new Date(ts * 1000).toLocaleTimeString("en-GB", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  });
}

const pad = (n: number) => String(n).padStart(2, "0");

/** Local "YYYY-MM-DD". */
export function dateKey(d: Date): string {
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

export function parseDateKey(k: string): Date {
  const [y, m, d] = k.split("-").map(Number);
  return new Date(y, m - 1, d);
}

export function shiftDate(k: string, days: number): string {
  const d = parseDateKey(k);
  d.setDate(d.getDate() + days);
  return dateKey(d);
}

/** "Tuesday, 6 Oct" */
export function prettyDate(k: string): string {
  return parseDateKey(k).toLocaleDateString("en-GB", {
    weekday: "long",
    day: "numeric",
    month: "short",
  });
}

export function pct(part: number, total: number): number {
  return total > 0 ? Math.round((part / total) * 100) : 0;
}
