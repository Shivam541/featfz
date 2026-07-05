import { proxyToBackend } from "@/lib/backend";

export async function POST(request: Request, context: { params: Promise<{ flagKey: string }> }) {
  const { flagKey } = await context.params;
  return proxyToBackend(request, `/v1/flags/${encodeURIComponent(flagKey)}/users/bulk-set`);
}

