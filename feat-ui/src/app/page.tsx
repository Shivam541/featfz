import Link from "next/link";

import { AppDetailsDisclosure } from "@/components/app-details-disclosure";
import { ModuleCard } from "@/components/module-card";

export default function Home() {
  return (
    <main className="min-h-screen px-4 py-5 text-[15px] text-[var(--foreground)] sm:px-6 lg:px-8">
      <div className="mx-auto flex min-h-[calc(100vh-2.5rem)] w-full max-w-[1180px] flex-col gap-4">
        <header className="glass-panel rounded-[var(--radius-xl)] px-5 py-4">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div className="max-w-2xl">
              <div className="flex items-center gap-4">
                <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-[var(--accent)] text-sm font-semibold text-white shadow-lg shadow-[rgba(10,107,109,0.25)]">
                  FZ
                </div>
                <div>
                  <h1 className="text-2xl font-semibold tracking-[-0.04em]">feat-ui</h1>
                  <p className="mt-1 text-sm text-[var(--muted)]">
                    Two focused modules: browse and evaluate flags, or manage their lifecycle.
                  </p>
                </div>
              </div>
            </div>

            <AppDetailsDisclosure />
          </div>
        </header>

        <section className="grid gap-4 lg:grid-cols-2">
          <ModuleCard
            title="Flags"
            description="List active flags, inspect a single flag, and run evaluation for one user."
            href="/flags"
            accent="View and evaluate"
            bullets={["Flag list", "Flag detail", "Evaluate one user"]}
          />
          <ModuleCard
            title="Manage"
            description="Create a new flag, update an existing one, or archive it when it is no longer needed."
            href="/manage"
            accent="Create and edit"
            bullets={["Create flag", "Update flag", "Archive flag"]}
          />
        </section>

        <section className="glass-panel rounded-[var(--radius-xl)] p-5">
          <div className="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <h2 className="text-lg font-semibold tracking-[-0.03em]">Navigation</h2>
              <p className="text-sm text-[var(--muted)]">
                Pick the module that matches what you want to do. The details stay out of the way until you expand them.
              </p>
            </div>
            <div className="flex gap-2">
              <Link
                href="/flags"
                className="rounded-2xl bg-[var(--accent)] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#0b7b7d]"
              >
                Open flags
              </Link>
              <Link
                href="/manage"
                className="rounded-2xl border border-[var(--border)] bg-white px-4 py-2.5 text-sm font-medium transition hover:border-[var(--accent)] hover:text-[var(--accent)]"
              >
                Open manage
              </Link>
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}

