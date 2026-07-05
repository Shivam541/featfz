import { proxyToBackend } from "@/lib/backend";

export async function GET(request: Request) {
  return proxyToBackend(request, "/v1/flags");
}

