import { defineEventHandler, getRequestURL, proxyRequest } from "h3";

export default defineEventHandler((event) => {
  const base = new URL(process.env.NUXT_BACKEND_URL || "http://localhost:8000");
  const url = getRequestURL(event);
  url.protocol = base.protocol;
  url.host = base.host;
  return proxyRequest(event, url.toString());
});
