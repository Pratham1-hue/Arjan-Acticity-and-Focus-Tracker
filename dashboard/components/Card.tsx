import type { Kind } from "@/lib/types";

export const KIND_COLOR: Record<Kind, string> = {
  productive: "#34d399",
  distraction: "#fb7185",
  neutral: "#71717a",
};

export function Card({
  title,
  action,
  children,
  className = "",
}: {
  title: string;
  action?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <section
      className={`rounded-2xl border border-zinc-800 bg-zinc-900/60 p-5 ${className}`}
    >
      <div className="mb-4 flex items-center justify-between gap-2">
        <h2 className="text-xs font-semibold uppercase tracking-wider text-zinc-500">
          {title}
        </h2>
        {action}
      </div>
      {children}
    </section>
  );
}

export function KindDot({ kind }: { kind: Kind }) {
  return (
    <span
      className="inline-block h-2 w-2 shrink-0 rounded-full"
      style={{ backgroundColor: KIND_COLOR[kind] }}
    />
  );
}

export function ShareBar({ pct, color }: { pct: number; color: string }) {
  return (
    <div className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-zinc-800">
      <div
        className="h-1.5 rounded-full transition-all duration-500"
        style={{ width: `${Math.max(2, Math.min(100, pct))}%`, backgroundColor: color }}
      />
    </div>
  );
}

export function Empty({ children }: { children: React.ReactNode }) {
  return <p className="text-sm text-zinc-500">{children}</p>;
}
