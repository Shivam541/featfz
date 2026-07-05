const encoder = new TextEncoder();

function base64UrlFromBytes(bytes: Uint8Array) {
  let binary = "";
  bytes.forEach((byte) => {
    binary += String.fromCharCode(byte);
  });
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
}

function base64UrlFromString(value: string) {
  return base64UrlFromBytes(encoder.encode(value));
}

export async function createHs256Jwt(input: {
  secret: string;
  appId: string;
  subject: string;
  expiresInSeconds: number;
}) {
  const now = Math.floor(Date.now() / 1000);
  const header = { alg: "HS256", typ: "JWT" };
  const payload = {
    app_id: input.appId.trim(),
    sub: input.subject.trim(),
    iat: now,
    exp: now + input.expiresInSeconds,
  };

  const signingInput = `${base64UrlFromString(JSON.stringify(header))}.${base64UrlFromString(
    JSON.stringify(payload),
  )}`;
  const key = await crypto.subtle.importKey(
    "raw",
    encoder.encode(input.secret),
    { name: "HMAC", hash: "SHA-256" },
    false,
    ["sign"],
  );
  const signature = new Uint8Array(await crypto.subtle.sign("HMAC", key, encoder.encode(signingInput)));
  return `${signingInput}.${base64UrlFromBytes(signature)}`;
}

