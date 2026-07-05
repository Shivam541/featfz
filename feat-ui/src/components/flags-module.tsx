"use client";

import { useEffect, useMemo, useState } from "react";

import { createHs256Jwt } from "@/lib/jwt";
import type { ErrorResponse, EvaluationResponse, Flag, FlagListResponse, FlagResponse } from "@/lib/types";

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

export function FlagsModule() {
  const [appId, setAppId] = useState(DEMO_APP_ID);
  const [secret, setSecret] = useState(DEMO_SECRET);
  const [subject, setSubject] = useState(DEMO_SUBJECT);
  const [expiresIn, setExpiresIn] = useState("3600");
  const [token, setToken] = useState("");
  const [flags, setFlags] = useState<Flag[]>([]);
  const [selectedKey, setSelectedKey] = useState("");
  const [selectedFlag, setSelectedFlag] = useState<Flag | null>(null);
  const [evalUserId, setEvalUserId] = useState("user_123");
  const [evaluation, setEvaluation] = useState("Awaiting a flag evaluation.");
  const [busy, setBusy] = useState<string | null>(null);
  const [feedback, setFeedback] = useState<string>("App details stay collapsed unless you expand them.");

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
      setFeedback("JWT minted locally. Load the flag list when ready.");
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

      setFlags(body.data.flags.filter((flag) => !flag.archived_at));
      const next = body.data.flags.find((flag) => !flag.archived_at) ?? null;
      if (next) {
        setSelectedKey(next.key);
        setSelectedFlag(next);
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
    } finally {
      setBusy(null);
    }
  }

  async function evaluateFlag() {
    if (!authenticated || !selectedKey || !evalUserId.trim()) return;

    setBusy("eval");
    try {
      const response = await fetch(
        `/api/eval?flag=${encodeURIComponent(selectedKey)}&user=${encodeURIComponent(evalUserId.trim())}`,
        {
          headers: {
            authorization: `Bearer ${token.trim()}`,
            "x-app-id": appId.trim(),
          },
          cache: "no-store",
        },
      );
      const body = (await readJson<EvaluationResponse>(response)) as ApiEnvelope<EvaluationResponse>;
      if (!response.ok || !("success" in body) || !body.success) {
        setEvaluation(parseError(body));
        return;
      }

      setEvaluation(`Flag ${selectedKey} is ${body.result.toUpperCase()} for ${evalUserId.trim()}.`);
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
          <label className="flex flex-col gap-1.5">
            <span className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">Bearer token</span>
            <textarea
              value={token}
              onChange={(event) => setToken(event.target.value)}
              rows={3}
              className="mono rounded-[var(--radius-md)] border border-[var(--border)] bg-white px-3 py-2.5 text-[13px] outline-none transition focus:border-[var(--accent)]"
            />
          </label>
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

        <div className="mt-5">
          <div className="mb-3 flex items-center justify-between">
            <h2 className="text-lg font-semibold tracking-[-0.03em]">Flags</h2>
            <span className="rounded-full border border-[var(--border)] bg-white px-3 py-1 text-xs font-medium">
              {flags.length}
            </span>
          </div>
          <div className="space-y-3">
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
                    {flag.default_enabled ? "On" : "Off"}
                  </span>
                </div>
                <div className="mt-3 text-xs text-[var(--muted)]">
                  Updated {formatTime(flag.updated_at)}
                </div>
              </button>
            ))}
            {flags.length === 0 ? (
              <div className="rounded-[var(--radius-lg)] border border-dashed border-[var(--border)] bg-white px-4 py-6 text-sm text-[var(--muted)]">
                Load the tenant flags to populate this list.
              </div>
            ) : null}
          </div>
        </div>
      </aside>

      <section className="grid gap-4 xl:grid-rows-[auto_auto]">
        <div className="glass-panel rounded-[var(--radius-xl)] p-5">
          <div className="flex flex-col gap-2 lg:flex-row lg:items-start lg:justify-between">
            <div>
              <h2 className="text-2xl font-semibold tracking-[-0.04em]">Flag detail</h2>
              <p className="mt-1 text-sm text-[var(--muted)]">
                Inspect the selected flag and run a one-user evaluation here.
              </p>
            </div>
            <div className="grid gap-2 sm:grid-cols-3">
              {[
                ["Selected", selectedFlag?.key || "None"],
                ["Default", selectedFlag ? (selectedFlag.default_enabled ? "On" : "Off") : "-"],
                ["Updated", selectedFlag ? formatTime(selectedFlag.updated_at) : "-"],
              ].map(([label, value]) => (
                <div key={label} className="rounded-[var(--radius-md)] border border-[var(--border)] bg-white px-3 py-2.5">
                  <div className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">{label}</div>
                  <div className="mt-1 text-sm font-medium">{value}</div>
                </div>
              ))}
            </div>
          </div>

          {selectedFlag ? (
            <div className="mt-5 grid gap-4 xl:grid-cols-[minmax(0,1.2fr)_minmax(320px,0.8fr)]">
              <div className="rounded-[var(--radius-lg)] border border-[var(--border)] bg-white p-4">
                <div className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">Description</div>
                <div className="mt-2 text-sm leading-6">{selectedFlag.description || "No description provided."}</div>
                <div className="mt-4 grid gap-3 sm:grid-cols-2">
                  <div className="rounded-[var(--radius-md)] border border-[var(--border)] bg-[var(--surface-soft)] px-3 py-2.5">
                    <div className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">Status</div>
                    <div className="mt-1 text-sm font-medium">{selectedFlag.default_enabled ? "Enabled by default" : "Disabled by default"}</div>
                  </div>
                  <div className="rounded-[var(--radius-md)] border border-[var(--border)] bg-[var(--surface-soft)] px-3 py-2.5">
                    <div className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">Archive</div>
                    <div className="mt-1 text-sm font-medium">{selectedFlag.archived_at ? "Archived" : "Active"}</div>
                  </div>
                </div>
              </div>

              <div className="rounded-[var(--radius-lg)] border border-[var(--border)] bg-white p-4">
                <div className="text-sm font-semibold">Evaluate</div>
                <p className="mt-1 text-sm text-[var(--muted)]">
                  Use the current tenant headers and check one user.
                </p>
                <label className="mt-4 flex flex-col gap-1.5">
                  <span className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">User ID</span>
                  <input
                    value={evalUserId}
                    onChange={(event) => setEvalUserId(event.target.value)}
                    className="rounded-[var(--radius-md)] border border-[var(--border)] px-3 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
                  />
                </label>
                <button
                  type="button"
                  onClick={() => void evaluateFlag()}
                  disabled={busy === "eval"}
                  className="mt-3 w-full rounded-2xl bg-[var(--accent)] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#0b7b7d] disabled:cursor-not-allowed disabled:opacity-70"
                >
                  {busy === "eval" ? "Evaluating..." : "Run evaluation"}
                </button>
                <div className="mt-4 rounded-[var(--radius-md)] border border-dashed border-[var(--border)] bg-[var(--surface-soft)] px-4 py-3 text-sm">
                  {evaluation}
                </div>
              </div>
            </div>
          ) : (
            <div className="mt-5 rounded-[var(--radius-lg)] border border-dashed border-[var(--border)] bg-[var(--surface-soft)] px-4 py-6 text-sm text-[var(--muted)]">
              Select a flag from the list to view its detail and evaluate it.
            </div>
          )}
        </div>
      </section>
    </section>
  );
}

