import crypto from "node:crypto";

const appId = process.env.APP_ID || "app-acme";
const secret = process.env.JWT_SECRET || "acme-secret";
const subject = process.env.JWT_SUBJECT || "smoke-user";
const now = Math.floor(Date.now() / 1000);
const expiresIn = Number(process.env.JWT_EXPIRES_IN || "3600");

const header = {
  alg: "HS256",
  typ: "JWT",
};

const payload = {
  app_id: appId,
  sub: subject,
  iat: now,
  exp: now + expiresIn,
};

function base64url(input) {
  return Buffer.from(JSON.stringify(input))
    .toString("base64")
    .replace(/=/g, "")
    .replace(/\+/g, "-")
    .replace(/\//g, "_");
}

const encodedHeader = base64url(header);
const encodedPayload = base64url(payload);
const signingInput = `${encodedHeader}.${encodedPayload}`;

const signature = crypto
  .createHmac("sha256", secret)
  .update(signingInput)
  .digest("base64")
  .replace(/=/g, "")
  .replace(/\+/g, "-")
  .replace(/\//g, "_");

const token = `${signingInput}.${signature}`;

console.log(token);
