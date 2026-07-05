import { proxyToBackend } from "@/lib/backend";

export async function GET(request: Request, context: { params: Promise<{ flagKey: string }> }) {
  const { flagKey } = await context.params;
  return proxyToBackend(request, `/v1/flags/${encodeURIComponent(flagKey)}`);
}

export async function PATCH(request: Request, context: { params: Promise<{ flagKey: string }> }) {
  const { flagKey } = await context.params;
  return proxyToBackend(request, `/v1/flags/${encodeURIComponent(flagKey)}`);
}

export async function DELETE(request: Request, context: { params: Promise<{ flagKey: string }> }) {
  const { flagKey } = await context.params;
  return proxyToBackend(request, `/v1/flags/${encodeURIComponent(flagKey)}`);
}

