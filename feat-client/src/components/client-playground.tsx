"use client";

import { useEffect, useMemo, useState } from "react";

import { createHs256Jwt } from "@/lib/jwt";
import type { AuthCheckResponse, ErrorResponse, EvaluationResponse } from "@/lib/types";

const DEMO_APP_ID = "app-acme";
const DEMO_SECRET = "acme-secret";
const DEMO_SUBJECT = "client-user";

type Mode = "auth" | "eval";

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

async function readJson<T>(response: Response): Promise<T | ErrorResponse> {
  try {
    return (await response.json()) as T | ErrorResponse;
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

function pretty(value: unknown) {
  return JSON.stringify(value, null, 2);
}

export function ClientPlayground() {
  const [appId, setAppId] = useState(DEMO_APP_ID);
  const [secret, setSecret] = useState(DEMO_SECRET);
  const [subject, setSubject] = useState(DEMO_SUBJECT);
  const [expiresIn, setExpiresIn] = useState("3600");
  const [flagKey, setFlagKey] = useState("new_dashboard");
  const [userId, setUserId] = useState("user_123");
  const [token, setToken] = useState("");
  const [mode, setMode] = useState<Mode>("eval");
  const [status, setStatus] = useState("Idle");
  const [responseBody, setResponseBody] = useState("No request has been sent yet.");
  const [curl, setCurl] = useState("");
  const [isBusy, setIsBusy] = useState(false);

  useEffect(() => {
    localStorage.setItem("feat-client-app-id", appId);
    localStorage.setItem("feat-client-secret", secret);
    localStorage.setItem("feat-client-subject", subject);
    localStorage.setItem("feat-client-expires", expiresIn);
    localStorage.setItem("feat-client-token", token);
    localStorage.setItem("feat-client-flag", flagKey);
    localStorage.setItem("feat-client-user", userId);
  }, [appId, expiresIn, flagKey, secret, subject, token, userId]);

  const authHeaders = useMemo(
    () => ({
      authorization: token.trim() ? `Bearer ${token.trim()}` : "",
      "x-app-id": appId.trim(),
    }),
    [appId, token],
  );

  async function mintToken() {
    const exp = Number.parseInt(expiresIn, 10);
    if (!Number.isFinite(exp) || exp < 60) {
      setStatus("Expiry must be at least 60 seconds.");
      return;
    }

    setStatus("Minting JWT locally...");
    try {
      const nextToken = await createHs256Jwt({
        secret,
        appId,
        subject,
        expiresInSeconds: exp,
      });
      setToken(nextToken);
      setStatus("JWT ready.");
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "JWT minting failed.");
    }
  }

  async function runRequest(nextMode: Mode) {
    if (!token.trim()) {
      setStatus("Mint a token before sending requests.");
      return;
    }

    setMode(nextMode);
    setIsBusy(true);

    try {
      if (nextMode === "auth") {
        const response = await fetch("/api/auth/check", {
          headers: authHeaders,
          cache: "no-store",
        });
        const body = (await readJson<AuthCheckResponse>(response)) as AuthCheckResponse | ErrorResponse;
        if (!response.ok || !("success" in body) || !body.success) {
          setStatus(parseError(body));
          setResponseBody(pretty(body));
          setCurl(`curl http://localhost:8080/v1/auth/check \\\n  -H "Authorization: Bearer $TOKEN" \\\n  -H "X-App-ID: ${appId.trim()}"`);
          return;
        }

        setStatus(`Auth check succeeded for ${body.data.app_id}.`);
        setResponseBody(pretty(body));
        setCurl(`curl http://localhost:8080/v1/auth/check \\\n  -H "Authorization: Bearer $TOKEN" \\\n  -H "X-App-ID: ${appId.trim()}"`);
        return;
      }

      const response = await fetch(
        `/api/eval?flag=${encodeURIComponent(flagKey.trim())}&user=${encodeURIComponent(userId.trim())}`,
        {
          headers: authHeaders,
          cache: "no-store",
        },
      );
      const body = (await readJson<EvaluationResponse>(response)) as EvaluationResponse | ErrorResponse;
      if (!response.ok || !("success" in body) || !body.success) {
        setStatus(parseError(body));
        setResponseBody(pretty(body));
        setCurl(
          `curl "http://localhost:8080/eval?flag=${encodeURIComponent(flagKey.trim())}&user=${encodeURIComponent(
            userId.trim(),
          )}" \\\n  -H "Authorization: Bearer $TOKEN" \\\n  -H "X-App-ID: ${appId.trim()}"`,
        );
        return;
      }

      setStatus(`Evaluation returned ${body.result.toUpperCase()} for ${userId.trim()}.`);
      setResponseBody(pretty(body));
      setCurl(
        `curl "http://localhost:8080/eval?flag=${encodeURIComponent(flagKey.trim())}&user=${encodeURIComponent(
          userId.trim(),
        )}" \\\n  -H "Authorization: Bearer $TOKEN" \\\n  -H "X-App-ID: ${appId.trim()}"`,
      );
    } finally {
      setIsBusy(false);
    }
  }

  return (
    <div className="min-h-screen px-4 py-5 text-[15px] text-[var(--foreground)] sm:px-6 lg:px-8">
      <div className="mx-auto flex min-h-[calc(100vh-2.5rem)] w-full max-w-[1400px] flex-col gap-4">
        <header className="glass-panel rounded-[var(--radius-xl)] px-5 py-4">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <div className="text-lg font-semibold tracking-[-0.03em]">feat-client</div>
              <p className="mt-1 max-w-2xl text-sm text-[var(--muted)]">
                Mint a tenant JWT, send auth or eval requests, and inspect the raw API response in one compact playground.
              </p>
            </div>
            <div className="flex flex-wrap gap-2 text-xs">
              <span className="rounded-full bg-[rgba(16,94,83,0.1)] px-3 py-1 font-medium text-[var(--accent)]">
                Local JWT minting
              </span>
              <span className="rounded-full bg-[rgba(18,122,79,0.1)] px-3 py-1 font-medium text-[var(--success)]">
                Backend proxy routes
              </span>
              <span className="rounded-full bg-[rgba(168,61,61,0.1)] px-3 py-1 font-medium text-[var(--danger)]">
                No shared tenant state
              </span>
            </div>
          </div>
        </header>

        <div className="grid flex-1 gap-4 xl:grid-cols-[minmax(0,0.92fr)_minmax(0,1.08fr)]">
          <section className="glass-panel rounded-[var(--radius-xl)] p-5">
            <h2 className="text-xl font-semibold tracking-[-0.03em]">JWT builder</h2>
            <p className="mt-1 text-sm text-[var(--muted)]">Generate a matching `app_id` token locally before hitting the service.</p>

            <div className="mt-4 grid gap-3 md:grid-cols-2">
              {[
                ["App ID", appId, setAppId],
                ["Secret", secret, setSecret],
                ["Subject", subject, setSubject],
                ["Expires in seconds", expiresIn, setExpiresIn],
              ].map(([label, value, setter]) => (
                <label key={label as string} className="flex flex-col gap-1.5">
                  <span className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">{label as string}</span>
                  <input
                    value={value as string}
                    onChange={(event) => (setter as (next: string) => void)(event.target.value)}
                    className="rounded-[var(--radius-md)] border border-[var(--border)] bg-white px-3 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
                  />
                </label>
              ))}
            </div>

            <button
              type="button"
              onClick={() => void mintToken()}
              className="mt-4 rounded-2xl bg-[var(--accent)] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#126c61]"
            >
              Mint token
            </button>

            <div className="mt-4 rounded-[var(--radius-lg)] border border-[var(--border)] bg-[var(--surface-soft)] p-4">
              <div className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">JWT preview</div>
              <div className="mono mt-2 break-all text-[13px] leading-5">{token || "Token will appear here."}</div>
            </div>

            <div className="mt-4 grid gap-3 md:grid-cols-2">
              <label className="flex flex-col gap-1.5">
                <span className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">Flag key</span>
                <input
                  value={flagKey}
                  onChange={(event) => setFlagKey(event.target.value)}
                  className="rounded-[var(--radius-md)] border border-[var(--border)] bg-white px-3 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
                />
              </label>
              <label className="flex flex-col gap-1.5">
                <span className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">User ID</span>
                <input
                  value={userId}
                  onChange={(event) => setUserId(event.target.value)}
                  className="rounded-[var(--radius-md)] border border-[var(--border)] bg-white px-3 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
                />
              </label>
            </div>

            <div className="mt-4 rounded-[var(--radius-lg)] border border-[var(--border)] bg-white p-4">
              <div className="text-sm font-semibold">Quick path</div>
              <p className="mt-1 text-sm text-[var(--muted)]">Use the toggle to choose between auth and eval, then run the request.</p>
              <div className="mt-3 grid grid-cols-2 gap-2">
                <button
                  type="button"
                  onClick={() => void runRequest("auth")}
                  disabled={isBusy}
                  className={`rounded-2xl border px-4 py-2.5 text-sm font-medium transition ${
                    mode === "auth"
                      ? "border-[var(--accent)] bg-[rgba(16,94,83,0.08)] text-[var(--accent)]"
                      : "border-[var(--border)] bg-[var(--surface-soft)]"
                  }`}
                >
                  Auth check
                </button>
                <button
                  type="button"
                  onClick={() => void runRequest("eval")}
                  disabled={isBusy}
                  className={`rounded-2xl border px-4 py-2.5 text-sm font-medium transition ${
                    mode === "eval"
                      ? "border-[var(--accent)] bg-[rgba(16,94,83,0.08)] text-[var(--accent)]"
                      : "border-[var(--border)] bg-[var(--surface-soft)]"
                  }`}
                >
                  Eval request
                </button>
              </div>
              <button
                type="button"
                onClick={() => void runRequest(mode)}
                disabled={isBusy}
                className="mt-3 w-full rounded-2xl bg-[var(--foreground)] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-black disabled:cursor-not-allowed disabled:opacity-70"
              >
                {isBusy ? "Running..." : "Send request"}
              </button>
            </div>
          </section>

          <section className="grid gap-4">
            <div className="glass-panel rounded-[var(--radius-xl)] p-5">
              <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
                <div>
                  <h2 className="text-xl font-semibold tracking-[-0.03em]">Response console</h2>
                  <p className="mt-1 text-sm text-[var(--muted)]">The exact JSON returned by the proxy route and backend.</p>
                </div>
                <div className="rounded-full bg-[rgba(16,94,83,0.1)] px-3 py-1 text-xs font-medium text-[var(--accent)]">
                  {status}
                </div>
              </div>
              <pre className="mono mt-4 overflow-auto rounded-[var(--radius-lg)] border border-[var(--border)] bg-[#0f1614] p-4 text-[13px] leading-6 text-[#d8f0ea]">
                {responseBody}
              </pre>
            </div>

            <div className="glass-panel rounded-[var(--radius-xl)] p-5">
              <h3 className="text-lg font-semibold tracking-[-0.03em]">Curl snippet</h3>
              <pre className="mono mt-3 overflow-auto rounded-[var(--radius-lg)] border border-[var(--border)] bg-white p-4 text-[13px] leading-6">
                {curl || "Run a request to generate a curl snippet."}
              </pre>
            </div>

            <div className="glass-panel rounded-[var(--radius-xl)] p-5">
              <h3 className="text-lg font-semibold tracking-[-0.03em]">Header recap</h3>
              <div className="mt-3 grid gap-3">
                <div className="rounded-[var(--radius-md)] border border-[var(--border)] bg-white px-4 py-3">
                  <div className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">Authorization</div>
                  <div className="mono mt-1 break-all text-sm">
                    {authHeaders.authorization || "Bearer <jwt>"}
                  </div>
                </div>
                <div className="rounded-[var(--radius-md)] border border-[var(--border)] bg-white px-4 py-3">
                  <div className="text-[11px] uppercase tracking-[0.16em] text-[var(--muted)]">X-App-ID</div>
                  <div className="mono mt-1 break-all text-sm">{authHeaders["x-app-id"] || "<app-id>"}</div>
                </div>
              </div>
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}
