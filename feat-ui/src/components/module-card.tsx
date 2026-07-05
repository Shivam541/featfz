import Link from "next/link";

type ModuleCardProps = {
  title: string;
  description: string;
  href: string;
  accent: string;
  bullets: string[];
};

export function ModuleCard({ title, description, href, accent, bullets }: ModuleCardProps) {
  return (
    <Link
      href={href}
      className="glass-panel group rounded-[var(--radius-xl)] p-6 transition hover:-translate-y-0.5 hover:shadow-[0_28px_100px_rgba(33,23,12,0.12)]"
    >
      <div className="flex items-start justify-between gap-4">
        <div>
          <div className="text-sm uppercase tracking-[0.2em] text-[var(--muted)]">{accent}</div>
          <h2 className="mt-2 text-2xl font-semibold tracking-[-0.04em]">{title}</h2>
          <p className="mt-2 max-w-xl text-sm leading-6 text-[var(--muted)]">{description}</p>
        </div>
        <div className="rounded-full border border-[var(--border)] bg-white px-3 py-1 text-xs font-medium">
          Open
        </div>
      </div>
      <div className="mt-6 grid gap-2 sm:grid-cols-3">
        {bullets.map((bullet) => (
          <div key={bullet} className="rounded-[var(--radius-md)] border border-[var(--border)] bg-white px-3 py-2 text-sm">
            {bullet}
          </div>
        ))}
      </div>
    </Link>
  );
}

