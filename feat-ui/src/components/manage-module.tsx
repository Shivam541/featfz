"use client";

import { useEffect, useMemo, useState } from "react";

import { createHs256Jwt } from "@/lib/jwt";
import type { ErrorResponse, Flag, FlagListResponse, FlagResponse } from "@/lib/types";

const DEMO_APP_ID = "app-acme";
const DEMO_SECRET = "acme-secret";
const DEMO_SUBJECT = "dashboard-user";

type ApiEnvelope<T> = T | ErrorResponse;

function parseError(body: unknown, fallback = "The request could not be completed.") {
  if (
    body &&
    typeof body === "object" &&
    "error" in body &&
    body.error &&
    typeof body.error === "object" &&
    "message" in body.error &&
    typeof body.error.message === "string"
  ) {
    return body.error.message;
  }
  return fallback;
}

async function readJson<T>(response: Response): Promise<ApiEnvelope<T>> {
  try {
    return (await response.json()) as ApiEnvelope<T>;
  } catch {
    return {
      success: false,
      error: {
        code: "invalid_response",
        message: "The service returned an invalid response.",
      },
    };
  }
}

function formatTime(value: string | null) {
  if (!value) return "Active";
  return new Date(value).toLocaleString([], {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

export function ManageModule() {
  const [appId, setAppId] = useState(DEMO_APP_ID);
  const [secret, setSecret] = useState(DEMO_SECRET);
  const [subject, setSubject] = useState(DEMO_SUBJECT);
  const [expiresIn, setExpiresIn] = useState("3600");
  const [token, setToken] = useState("");
  const [flags, setFlags] = useState<Flag[]>([]);
  const [selectedKey, setSelectedKey] = useState("");
  const [selectedFlag, setSelectedFlag] = useState<Flag | null>(null);
  const [busy, setBusy] = useState<string | null>(null);
  const [feedback, setFeedback] = useState("Create and edit controls stay in this module only.");
  const [draft, setDraft] = useState({
    key: "",
    description: "",
    default_enabled: false,
  });
  const [updateDraft, setUpdateDraft] = useState({
    description: "",
    default_enabled: false,
  });

  const authenticated = useMemo(() => token.trim().length > 0 && appId.trim().length > 0, [appId, token]);

  useEffect(() => {
    localStorage.setItem("feat-ui-app-id", appId);
    localStorage.setItem("feat-ui-secret", secret);
    localStorage.setItem("feat-ui-subject", subject);
    localStorage.setItem("feat-ui-expires", expiresIn);
    localStorage.setItem("feat-ui-token", token);
  }, [appId, expiresIn, secret, subject, token]);

  async function mintToken() {
    const exp = Number.parseInt(expiresIn, 10);
    if (!Number.isFinite(exp) || exp < 60) {
      setFeedback("Expiry must be at least 60 seconds.");
      return;
    }

    setBusy("mint");
    try {
      const nextToken = await createHs256Jwt({
        secret,
        appId,
        subject,
        expiresInSeconds: exp,
      });
      setToken(nextToken);
      setFeedback("JWT minted locally. You can now manage flags.");
    } finally {
      setBusy(null);
    }
  }

  async function loadFlags() {
    if (!authenticated) {
      setFeedback("Mint a JWT first.");
      return;
    }

    setBusy("load");
    try {
      const response = await fetch("/api/flags", {
        headers: {
          authorization: `Bearer ${token.trim()}`,
          "x-app-id": appId.trim(),
        },
        cache: "no-store",
      });
      const body = (await readJson<FlagListResponse>(response)) as ApiEnvelope<FlagListResponse>;
      if (!response.ok || !("success" in body) || !body.success) {
        setFeedback(parseError(body));
        return;
      }

      setFlags(body.data.flags);
      const next = body.data.flags.find((flag) => flag.key === selectedKey) ?? body.data.flags[0] ?? null;
      if (next) {
        setSelectedKey(next.key);
        setSelectedFlag(next);
        setUpdateDraft({
          description: next.description,
          default_enabled: next.default_enabled,
        });
      }
      setFeedback(`Loaded ${body.data.flags.length} flag${body.data.flags.length === 1 ? "" : "s"}.`);
    } finally {
      setBusy(null);
    }
  }

  async function selectFlag(flagKey: string) {
    if (!authenticated) return;

    setBusy(`flag-${flagKey}`);
    try {
      const response = await fetch(`/api/flags/${encodeURIComponent(flagKey)}`, {
        headers: {
          authorization: `Bearer ${token.trim()}`,
          "x-app-id": appId.trim(),
        },
        cache: "no-store",
      });
      const body = (await readJson<FlagResponse>(response)) as ApiEnvelope<FlagResponse>;
      if (!response.ok || !("success" in body) || !body.success) {
        setFeedback(parseError(body));
        return;
      }

      setSelectedKey(flagKey);
      setSelectedFlag(body.data);
      setUpdateDraft({
        description: body.data.description,
        default_enabled: body.data.default_enabled,
      });
    } finally {
      setBusy(null);
    }
  }

  async function createFlag() {
    if (!authenticated) return;

    setBusy("create");
    try {
      const response = await fetch("/api/flags", {
        method: "POST",
        headers: {
          authorization: `Bearer ${token.trim()}`,
          "x-app-id": appId.trim(),
          "content-type": "application/json",
        },
        body: JSON.stringify(draft),
      });
      const body = (await readJson<FlagResponse>(response)) as ApiEnvelope<FlagResponse>;
      if (!response.ok || !("success" in body) || !body.success) {
        setFeedback(parseError(body));
        return;
      }

      setDraft({ key: "", description: "", default_enabled: false });
      await loadFlags();
      setSelectedKey(body.data.key);
      setSelectedFlag(body.data);
      setFeedback(`Created ${body.data.key}.`);
    } finally {
      setBusy(null);
    }
  }

  async function updateFlag() {
    if (!authenticated || !selectedKey) return;

    setBusy("update");
    try {
      const response = await fetch(`/api/flags/${encodeURIComponent(selectedKey)}`, {
        method: "PATCH",
        headers: {
          authorization: `Bearer ${token.trim()}`,
          "x-app-id": appId.trim(),
          "content-type": "application/json",
        },
        body: JSON.stringify(updateDraft),
      });
      const body = (await readJson<FlagResponse>(response)) as ApiEnvelope<FlagResponse>;
      if (!response.ok || !("success" in body) || !body.success) {
        setFeedback(parseError(body));
        return;
      }

      setSelectedFlag(body.data);
      await loadFlags();
      setFeedback(`Updated ${body.data.key}.`);
    } finally {
      setBusy(null);
    }
  }

  async function archiveFlag() {
    if (!authenticated || !selectedKey) return;

    setBusy("delete");
    try {
      const response = await fetch(`/api/flags/${encodeURIComponent(selectedKey)}`, {
        method: "DELETE",
        headers: {
          authorization: `Bearer ${token.trim()}`,
          "x-app-id": appId.trim(),
        },
      });
      const body = (await readJson<{ success: true }>(response)) as ApiEnvelope<{ success: true }>;
      if (!response.ok || !("success" in body) || !body.success) {
        setFeedback(parseError(body));
        return;
      }

      setSelectedFlag((current) => (current ? { ...current, archived_at: new Date().toISOString() } : current));
      await loadFlags();
      setFeedback(`Archived ${selectedKey}.`);
    } finally {
      setBusy(null);
    }
  }

  return (
    <section className="grid gap-4 xl:grid-cols-[360px_minmax(0,1fr)]">
      <aside className="glass-panel rounded-[var(--radius-xl)] p-5">
        <div className="space-y-4">
          <label className="flex flex-col gap-1.5">
            <span className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">App ID</span>
            <input
              value={appId}
              onChange={(event) => setAppId(event.target.value)}
              className="rounded-[var(--radius-md)] border border-[var(--border)] bg-white px-3 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
            />
          </label>
          <label className="flex flex-col gap-1.5">
            <span className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">Secret</span>
            <input
              value={secret}
              onChange={(event) => setSecret(event.target.value)}
              className="rounded-[var(--radius-md)] border border-[var(--border)] bg-white px-3 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
            />
          </label>
          <label className="flex flex-col gap-1.5">
            <span className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">Subject</span>
            <input
              value={subject}
              onChange={(event) => setSubject(event.target.value)}
              className="rounded-[var(--radius-md)] border border-[var(--border)] bg-white px-3 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
            />
          </label>
          <label className="flex flex-col gap-1.5">
            <span className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">Expires in seconds</span>
            <input
              value={expiresIn}
              onChange={(event) => setExpiresIn(event.target.value)}
              className="rounded-[var(--radius-md)] border border-[var(--border)] bg-white px-3 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
            />
          </label>
          <button
            type="button"
            onClick={() => void mintToken()}
            disabled={busy === "mint"}
            className="w-full rounded-2xl bg-[var(--accent)] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#0b7b7d] disabled:cursor-not-allowed disabled:opacity-70"
          >
            {busy === "mint" ? "Minting..." : "Mint token"}
          </button>
          <button
            type="button"
            onClick={() => void loadFlags()}
            disabled={busy === "load"}
            className="w-full rounded-2xl border border-[var(--border)] bg-white px-4 py-2.5 text-sm font-medium transition hover:border-[var(--accent)] hover:text-[var(--accent)] disabled:cursor-not-allowed disabled:opacity-70"
          >
            {busy === "load" ? "Loading..." : "Load flags"}
          </button>
          <div className="rounded-[var(--radius-lg)] border border-[var(--border)] bg-[var(--surface-soft)] px-4 py-3 text-sm text-[var(--muted)]">
            {feedback}
          </div>
        </div>

        <div className="mt-5 space-y-3">
          <h2 className="text-lg font-semibold tracking-[-0.03em]">Existing flags</h2>
          {flags.map((flag) => (
            <button
              key={flag.key}
              type="button"
              onClick={() => void selectFlag(flag.key)}
              className={`w-full rounded-[var(--radius-lg)] border p-4 text-left transition ${
                selectedKey === flag.key
                  ? "border-[var(--accent)] bg-white shadow-lg shadow-[rgba(10,107,109,0.12)]"
                  : "border-[var(--border)] bg-white hover:border-[rgba(10,107,109,0.25)]"
              }`}
            >
              <div className="flex items-start justify-between gap-3">
                <div>
                  <div className="font-semibold">{flag.key}</div>
                  <div className="mt-1 text-sm text-[var(--muted)]">{flag.description || "No description"}</div>
                </div>
                <span className="rounded-full bg-[rgba(16,94,83,0.1)] px-2.5 py-1 text-[11px] font-medium uppercase tracking-[0.14em] text-[var(--accent)]">
                  {flag.archived_at ? "Archived" : flag.default_enabled ? "On" : "Off"}
                </span>
              </div>
            </button>
          ))}
          {flags.length === 0 ? (
            <div className="rounded-[var(--radius-lg)] border border-dashed border-[var(--border)] bg-white px-4 py-6 text-sm text-[var(--muted)]">
              Load flags to choose one to manage.
            </div>
          ) : null}
        </div>
      </aside>

      <section className="glass-panel rounded-[var(--radius-xl)] p-5">
        <div className="flex flex-col gap-2 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <h2 className="text-2xl font-semibold tracking-[-0.04em]">Manage flag</h2>
            <p className="mt-1 text-sm text-[var(--muted)]">
              Create a new flag or adjust the currently selected one.
            </p>
          </div>
          <div className="grid gap-2 sm:grid-cols-3">
            {[
              ["Selected", selectedFlag?.key || "None"],
              ["Status", selectedFlag ? (selectedFlag.archived_at ? "Archived" : "Active") : "-"],
              ["Updated", selectedFlag ? formatTime(selectedFlag.updated_at) : "-"],
            ].map(([label, value]) => (
              <div key={label} className="rounded-[var(--radius-md)] border border-[var(--border)] bg-white px-3 py-2.5">
                <div className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">{label}</div>
                <div className="mt-1 text-sm font-medium">{value}</div>
              </div>
            ))}
          </div>
        </div>

        <div className="mt-5 grid gap-4 xl:grid-cols-2">
          <div className="rounded-[var(--radius-lg)] border border-[var(--border)] bg-white p-4">
            <h3 className="text-sm font-semibold uppercase tracking-[0.16em] text-[var(--muted)]">Create</h3>
            <div className="mt-4 space-y-3">
              <label className="flex flex-col gap-1.5">
                <span className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">Key</span>
                <input
                  value={draft.key}
                  onChange={(event) => setDraft((current) => ({ ...current, key: event.target.value }))}
                  className="rounded-[var(--radius-md)] border border-[var(--border)] px-3 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
                />
              </label>
              <label className="flex flex-col gap-1.5">
                <span className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">Description</span>
                <textarea
                  rows={4}
                  value={draft.description}
                  onChange={(event) => setDraft((current) => ({ ...current, description: event.target.value }))}
                  className="rounded-[var(--radius-md)] border border-[var(--border)] px-3 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
                />
              </label>
              <label className="flex items-center justify-between rounded-[var(--radius-md)] border border-[var(--border)] px-3 py-3">
                <div>
                  <div className="text-sm font-medium">Default enabled</div>
                  <div className="text-xs text-[var(--muted)]">Fallback for evaluations.</div>
                </div>
                <input
                  type="checkbox"
                  checked={draft.default_enabled}
                  onChange={(event) => setDraft((current) => ({ ...current, default_enabled: event.target.checked }))}
                  className="h-4 w-4 accent-[var(--accent)]"
                />
              </label>
              <button
                type="button"
                onClick={() => void createFlag()}
                disabled={busy === "create"}
                className="rounded-2xl bg-[var(--accent)] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#0b7b7d] disabled:cursor-not-allowed disabled:opacity-70"
              >
                {busy === "create" ? "Creating..." : "Create flag"}
              </button>
            </div>
          </div>

          <div className="rounded-[var(--radius-lg)] border border-[var(--border)] bg-[var(--surface-soft)] p-4">
            <h3 className="text-sm font-semibold uppercase tracking-[0.16em] text-[var(--muted)]">Update / archive</h3>
            {selectedFlag ? (
              <div className="mt-4 space-y-3">
                <label className="flex flex-col gap-1.5">
                  <span className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">Description</span>
                  <textarea
                    rows={4}
                    value={updateDraft.description}
                    onChange={(event) =>
                      setUpdateDraft((current) => ({ ...current, description: event.target.value }))
                    }
                    className="rounded-[var(--radius-md)] border border-[var(--border)] bg-white px-3 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
                  />
                </label>
                <label className="flex items-center justify-between rounded-[var(--radius-md)] border border-[var(--border)] bg-white px-3 py-3">
                  <div>
                    <div className="text-sm font-medium">Default enabled</div>
                    <div className="text-xs text-[var(--muted)]">This flips the fallback state.</div>
                  </div>
                  <input
                    type="checkbox"
                    checked={updateDraft.default_enabled}
                    onChange={(event) =>
                      setUpdateDraft((current) => ({ ...current, default_enabled: event.target.checked }))
                    }
                    className="h-4 w-4 accent-[var(--accent)]"
                  />
                </label>
                <div className="flex flex-wrap gap-3">
                  <button
                    type="button"
                    onClick={() => void updateFlag()}
                    disabled={busy === "update"}
                    className="rounded-2xl bg-[var(--foreground)] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-black disabled:cursor-not-allowed disabled:opacity-70"
                  >
                    {busy === "update" ? "Saving..." : "Update flag"}
                  </button>
                  <button
                    type="button"
                    onClick={() => void archiveFlag()}
                    disabled={busy === "delete"}
                    className="rounded-2xl border border-[rgba(168,61,61,0.22)] bg-[rgba(168,61,61,0.06)] px-4 py-2.5 text-sm font-medium text-[var(--danger)] transition hover:bg-[rgba(168,61,61,0.1)] disabled:cursor-not-allowed disabled:opacity-70"
                  >
                    {busy === "delete" ? "Archiving..." : "Archive flag"}
                  </button>
                </div>
              </div>
            ) : (
              <div className="mt-4 rounded-[var(--radius-md)] border border-dashed border-[var(--border)] bg-white px-4 py-6 text-sm text-[var(--muted)]">
                Select a flag from the left or create a new one above.
              </div>
            )}
          </div>
        </div>
      </section>
    </section>
  );
}

