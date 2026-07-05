import { proxyToBackend } from "@/lib/backend";

export async function GET(request: Request) {
  const url = new URL(request.url);
  return proxyToBackend(request, `/eval${url.search}`);
}

