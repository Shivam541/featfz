import Link from "next/link";

import { AppDetailsDisclosure } from "@/components/app-details-disclosure";
import { FlagsModule } from "@/components/flags-module";

export default function FlagsPage() {
  return (
    <main className="min-h-screen px-4 py-5 text-[15px] text-[var(--foreground)] sm:px-6 lg:px-8">
      <div className="mx-auto flex min-h-[calc(100vh-2.5rem)] w-full max-w-[1440px] flex-col gap-4">
        <header className="glass-panel rounded-[var(--radius-xl)] px-5 py-4">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div>
              <div className="text-sm uppercase tracking-[0.2em] text-[var(--muted)]">Module</div>
              <h1 className="text-2xl font-semibold tracking-[-0.04em]">Flags</h1>
              <p className="mt-1 max-w-2xl text-sm text-[var(--muted)]">
                View the tenant flag list, inspect a flag, and evaluate it for a single user on the same page.
              </p>
            </div>
            <div className="flex flex-wrap items-start gap-2">
              <Link
                href="/"
                className="rounded-2xl border border-[var(--border)] bg-white px-4 py-2.5 text-sm font-medium transition hover:border-[var(--accent)] hover:text-[var(--accent)]"
              >
                Home
              </Link>
              <Link
                href="/manage"
                className="rounded-2xl bg-[var(--accent)] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#0b7b7d]"
              >
                Go to manage
              </Link>
              <AppDetailsDisclosure />
            </div>
          </div>
        </header>

        <FlagsModule />
      </div>
    </main>
  );
}

