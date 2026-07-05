"use client";

const rows = [
  ["App ID", "app-acme"],
  ["Secret", "acme-secret"],
  ["Subject", "dashboard-user"],
  ["Expires", "3600"],
];

export function AppDetailsDisclosure() {
  return (
    <details className="group rounded-[var(--radius-lg)] border border-[var(--border)] bg-white px-4 py-3">
      <summary className="cursor-pointer list-none text-sm font-medium text-[var(--foreground)]">
        App details
        <span className="ml-2 text-[11px] uppercase tracking-[0.16em] text-[var(--muted)] group-open:hidden">
          Expand
        </span>
        <span className="ml-2 hidden text-[11px] uppercase tracking-[0.16em] text-[var(--muted)] group-open:inline">
          Collapse
        </span>
      </summary>
      <div className="mt-3 grid gap-2 sm:grid-cols-2 lg:min-w-[560px]">
        {rows.map(([label, value]) => (
          <div key={label} className="rounded-[var(--radius-md)] border border-[var(--border)] bg-[var(--surface-soft)] px-3 py-2">
            <div className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">{label}</div>
            <div className="mt-1 text-sm font-medium">{value}</div>
          </div>
        ))}
      </div>
    </details>
  );
}

