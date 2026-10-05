// SVG progress ring for the Focus Score.

interface FocusRingProps {
  score: number | null;
}

export function FocusRing({ score }: FocusRingProps) {
  const r = 52;
  const circ = 2 * Math.PI * r;
  const pct = score === null ? 0 : score / 100;
  const color =
    score === null
      ? "#52525b"
      : score >= 70
        ? "#34d399" // emerald
        : score >= 40
          ? "#fbbf24" // amber
          : "#fb7185"; // rose

  return (
    <div className="relative mx-auto h-40 w-40">
      <svg viewBox="0 0 120 120" className="h-full w-full -rotate-90">
        <circle
          cx="60"
          cy="60"
          r={r}
          fill="none"
          stroke="#27272a"
          strokeWidth="10"
        />
        <circle
          cx="60"
          cy="60"
          r={r}
          fill="none"
          stroke={color}
          strokeWidth="10"
          strokeLinecap="round"
          strokeDasharray={circ}
          strokeDashoffset={circ * (1 - pct)}
          className="transition-all duration-700"
        />
      </svg>
      <div className="absolute inset-0 flex flex-col items-center justify-center">
        <span
          className="text-4xl font-bold tabular-nums"
          style={{ color }}
        >
          {score === null ? "—" : score}
        </span>
        <span className="mt-1 text-[10px] font-semibold tracking-widest text-zinc-500">
          FOCUS SCORE
        </span>
      </div>
    </div>
  );
}
