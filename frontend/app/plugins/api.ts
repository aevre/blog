import { appendResponseHeader } from "h3";

export default defineNuxtPlugin(() => {
  const event = import.meta.server ? useRequestEvent() : undefined;
  const cookies = new Map(
    (event?.node.req.headers.cookie ?? "").split(";").flatMap((item) => {
      const index = item.indexOf("=");
      return index < 0 ? [] : [[item.slice(0, index).trim(), item.slice(index + 1)] as const];
    }),
  );

  const api = $fetch.create({
    baseURL: useRuntimeConfig().public.apiBase,

    onRequest({ options }) {
      if (!import.meta.server) return;
      const cookie = [...cookies].map(([name, value]) => `${name}=${value}`).join("; ");

      if (cookie) {
        options.headers.set("cookie", cookie);
      }
    },

    onResponse({ response }) {
      if (!event) return;

      for (const cookie of response.headers.getSetCookie()) {
        appendResponseHeader(event, "set-cookie", cookie);
        const pair = cookie.split(";", 1)[0]!;
        const index = pair.indexOf("=");
        if (index > 0) {
          cookies.set(pair.slice(0, index).trim(), pair.slice(index + 1));
        }
      }
    },
  });

  return {
    provide: {
      api,
    },
  };
});
