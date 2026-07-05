const backendBaseUrl = process.env.BACKEND_BASE_URL ?? "http://127.0.0.1:8080";

function forwardHeaders(headers: Headers) {
  const forwarded = new Headers();
  for (const name of ["authorization", "x-app-id", "content-type", "accept"]) {
    const value = headers.get(name);
    if (value) {
      forwarded.set(name, value);
    }
  }
  return forwarded;
}

export async function proxyToBackend(
  request: Request,
  targetPath: string,
  init?: RequestInit,
) {
  const url = new URL(targetPath, backendBaseUrl);
  const method = init?.method ?? request.method;
  const headers = new Headers(init?.headers ?? forwardHeaders(request.headers));

  const body =
    method === "GET" || method === "HEAD" ? undefined : await request.arrayBuffer();

  const upstream = await fetch(url, {
    method,
    headers,
    body,
    cache: "no-store",
  });

  return new Response(upstream.body, {
    status: upstream.status,
    headers: upstream.headers,
  });
}

