import { createHash } from "node:crypto";
import type { H3Event } from "h3";

const API_PREFIX = "/api";

const ACCESS_TOKEN = "accessToken";
const REFRESH_TOKEN = "refreshToken";

const RESPONSE_HEADERS = [
  "content-type",
  "cache-control",
  "etag",
  "last-modified",
  "location",
  "content-disposition",
];

const REQUEST_HEADERS = ["content-type", "accept"];

const EXCLUDED_REFRESH_PATHS = new Set(["auth/login", "auth/logout", "auth/refresh"]);

type RefreshResult = {
  accessToken: string;
  refreshToken?: string;
  cookies: string[];
};

const refreshPromises = new Map<string, Promise<RefreshResult | null>>();

/**
 * Resolve and validate the upstream API URL.
 */
function getUpstreamURL(event: H3Event, path: string): URL {
  const { apiBase } = useRuntimeConfig(event);

  if (!apiBase) {
    throw createError({
      statusCode: 500,
      statusMessage: "API base URL is not configured",
    });
  }

  const base = new URL(`${apiBase.replace(/\/+$/, "")}/`);
  const suffix = path.replace(/^\/+/, "");

  for (const segment of suffix.split("/")) {
    let decoded: string;

    try {
      decoded = decodeURIComponent(segment);
    } catch {
      throw createError({
        statusCode: 400,
        statusMessage: "Invalid API path",
      });
    }

    if (
      decoded === "." ||
      decoded === ".." ||
      decoded.includes("/") ||
      decoded.includes("\\") ||
      decoded.includes("%")
    ) {
      throw createError({
        statusCode: 400,
        statusMessage: "Invalid API path",
      });
    }
  }

  const url = new URL(suffix, base);

  if (url.origin !== base.origin || !url.pathname.startsWith(base.pathname)) {
    throw createError({
      statusCode: 400,
      statusMessage: "Invalid API path",
    });
  }

  return url;
}

/**
 * Extract all Set-Cookie headers.
 */
function getResponseCookies(response: Response): string[] {
  return response.headers.getSetCookie();
}

/**
 * Extract a cookie value from Set-Cookie headers.
 */
function getTokenFromCookies(cookies: string[], name: string): string | undefined {
  for (const cookie of cookies) {
    const [pair] = cookie.split(";");

    if (!pair) continue;

    const index = pair.indexOf("=");

    if (index === -1) continue;

    const key = pair.slice(0, index).trim();
    const value = pair.slice(index + 1);

    if (key === name) {
      return value;
    }
  }

  return undefined;
}

/**
 * Forward Set-Cookie headers to the client.
 */
function forwardCookies(event: H3Event, cookies: string[]): void {
  for (const cookie of cookies) {
    appendResponseHeader(event, "set-cookie", cookie);
  }
}

/**
 * Forward selected upstream response headers.
 */
function forwardResponseHeaders(event: H3Event, response: Response): void {
  for (const name of RESPONSE_HEADERS) {
    const value = response.headers.get(name);

    if (value !== null) {
      setResponseHeader(event, name, value);
    }
  }
}

/**
 * Replace cookie values in an existing Cookie header.
 */
function updateCookieHeader(
  original: string | undefined,
  tokens: Record<string, string | undefined>,
): string {
  const cookies = new Map<string, string>();

  for (const item of (original ?? "").split(";")) {
    const index = item.indexOf("=");

    if (index === -1) continue;

    const name = item.slice(0, index).trim();
    const value = item.slice(index + 1).trim();

    if (name) {
      cookies.set(name, value);
    }
  }

  for (const [name, value] of Object.entries(tokens)) {
    if (value !== undefined) {
      cookies.set(name, value);
    }
  }

  return [...cookies].map(([name, value]) => `${name}=${value}`).join("; ");
}

/**
 * Create a stable key without storing the raw refresh token.
 */
function getRefreshKey(token: string): string {
  return createHash("sha256").update(token).digest("hex");
}

/**
 * Share an in-flight refresh operation among concurrent requests.
 */
function refreshOnce(
  key: string,
  refresh: () => Promise<RefreshResult | null>,
): Promise<RefreshResult | null> {
  const existing = refreshPromises.get(key);

  if (existing) {
    return existing;
  }

  const promise = Promise.resolve().then(refresh);

  refreshPromises.set(key, promise);

  void promise.then(
    () => {
      if (refreshPromises.get(key) === promise) {
        refreshPromises.delete(key);
      }
    },
    () => {
      if (refreshPromises.get(key) === promise) {
        refreshPromises.delete(key);
      }
    },
  );

  return promise;
}

/**
 * Refresh authentication tokens through the Go backend.
 */
async function refreshTokens(event: H3Event, refreshToken: string): Promise<RefreshResult | null> {
  const key = getRefreshKey(refreshToken);

  return refreshOnce(key, async () => {
    const headers = new Headers({
      cookie: `${REFRESH_TOKEN}=${encodeURIComponent(refreshToken)}`,
    });

    const response = await fetch(getUpstreamURL(event, "auth/refresh"), {
      method: "POST",
      headers,
      redirect: "manual",
      signal: AbortSignal.timeout(10_000),
    });

    if (!response.ok) {
      return null;
    }

    const cookies = getResponseCookies(response);

    const rawAccessToken = getTokenFromCookies(cookies, ACCESS_TOKEN);

    if (!rawAccessToken) {
      return null;
    }

    const rawRefreshToken = getTokenFromCookies(cookies, REFRESH_TOKEN);

    return {
      accessToken: decodeURIComponent(rawAccessToken),
      refreshToken: rawRefreshToken,
      cookies,
    };
  });
}

/**
 * Proxy API requests and refresh expired access tokens.
 */
export default defineEventHandler(async (event) => {
  const requestURL = getRequestURL(event);
  const pathname = requestURL.pathname;

  if (pathname !== API_PREFIX && !pathname.startsWith(`${API_PREFIX}/`)) {
    throw createError({
      statusCode: 404,
      statusMessage: "API route not found",
    });
  }

  const path = pathname.slice(API_PREFIX.length).replace(/^\/+/, "");

  const normalizedPath = path.replace(/\/+$/, "");
  const method = event.method.toUpperCase();

  const url = getUpstreamURL(event, path);
  url.search = requestURL.search;

  // Buffer the request body so it can be replayed after a refresh.
  const rawBody = ["GET", "HEAD"].includes(method) ? undefined : await readRawBody(event, false);

  const body = rawBody ? Uint8Array.from(rawBody).buffer : undefined;

  const headers = new Headers();

  for (const name of REQUEST_HEADERS) {
    const value = getRequestHeader(event, name);

    if (value !== undefined) {
      headers.set(name, value);
    }
  }

  const originalCookie = getRequestHeader(event, "cookie");

  if (originalCookie) {
    headers.set("cookie", originalCookie);
  }

  const accessToken = getCookie(event, ACCESS_TOKEN);

  if (accessToken) {
    headers.set("authorization", `Bearer ${accessToken}`);
  }

  /**
   * Execute the upstream request with the supplied headers.
   */
  const request = (requestHeaders: Headers): Promise<Response> => {
    return fetch(url, {
      method,
      headers: requestHeaders,
      body,
      redirect: "manual",
      signal: AbortSignal.timeout(30_000),
    });
  };

  let response: Response;

  try {
    response = await request(headers);
  } catch {
    throw createError({
      statusCode: 502,
      statusMessage: "Backend API unavailable",
    });
  }

  const shouldRefresh = response.status === 401 && !EXCLUDED_REFRESH_PATHS.has(normalizedPath);

  if (shouldRefresh) {
    const refreshToken = getCookie(event, REFRESH_TOKEN);

    if (refreshToken) {
      try {
        const result = await refreshTokens(event, refreshToken);

        if (result) {
          headers.set("authorization", `Bearer ${result.accessToken}`);

          headers.set(
            "cookie",
            updateCookieHeader(originalCookie, {
              [ACCESS_TOKEN]: encodeURIComponent(result.accessToken),
              [REFRESH_TOKEN]: result.refreshToken,
            }),
          );

          // Retry the original request once.
          const retryResponse = await request(headers);

          // Forward refreshed cookies only after a completed retry.
          forwardCookies(event, result.cookies);

          response = retryResponse;
        }
      } catch {
        // Preserve the original response if refresh or retry fails.
      }
    }
  }

  setResponseStatus(event, response.status);

  forwardResponseHeaders(event, response);

  forwardCookies(event, getResponseCookies(response));

  if (method === "HEAD") {
    return null;
  }

  return response.body;
});
