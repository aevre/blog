import { ofetch, type FetchError, type FetchOptions } from "ofetch";
import type { ApiResponse } from "~/types/api";

type RefreshData = {
  accessToken: string;
};
type FetchBody = BodyInit | Record<string, unknown> | null | undefined;

/**
 * Flag indicating whether the Access Token is currently being refreshed
 */
let isRefreshing = false;

/**
 * Queue of pending requests while a token refresh is in progress
 */
const pendingRequests: Array<(err?: Error) => void> = [];

/**
 * Base ofetch instance with default baseURL and JSON headers
 */
const $baseFetch = ofetch.create({
  baseURL: "/api/v1",
  headers: {
    "Content-Type": "application/json",
  },
});

/**
 * Client-side API composable that provides a fetch function with automatic
 * token handling and refresh logic.
 */
export function useClientApi() {
  async function doRequest<T>(
    url: string,
    options: FetchOptions<"json", unknown> = {},
  ): Promise<T> {
    const accessToken = useAuthStore().accessToken;
    const headers = accessToken
      ? { ...(options.headers || {}), Authorization: `Bearer ${accessToken}` }
      : options.headers;
    return $baseFetch<T>(url, { ...options, headers });
  }
  /**
   * Client-only API request function. Automatically attaches the Access Token
   * and supports token refresh.
   *
   * **Note: This function must only be called on the client.**
   * Calling it during SSR will throw an error.
   *
   * @template T - Type of the returned data
   * @param {string} url - The URL to request (relative or absolute)
   * @param {FetchOptions<"json", unknown>} [options={}] - Optional request configuration
   * @throws {Error} Throws "apiFetch must be called on the client" if called during SSR
   * @returns {Promise<T>} Returns a promise resolving to the data of type T
   *
   * @example
   * import { useClientApi } from '~/composables/useApi'
   *
   * const { apiFetch } = useClientApi()
   *
   * onMounted(async () => {
   *   const user = await apiFetch<User>('/api/protected/user', { method: 'get' })
   *   console.log(user)
   * })
   */
  async function apiFetch<T>(url: string, options: FetchOptions<"json", unknown> = {}): Promise<T> {
    if (import.meta.server) {
      throw new Error("apiFetch must be called on the client");
    }

    try {
      return await doRequest<T>(url, options);
    } catch (err: unknown) {
      const fetchErr = err as FetchError;

      if (fetchErr?.response?.status !== 401) {
        throw fetchErr;
      }

      try {
        await refreshAccessToken();
      } catch (refreshErr) {
        useAuthStore().logout();
        await navigateTo("/auth/login");
        throw refreshErr;
      }

      return doRequest<T>(url, options);
    }
  }

  /**
   * Refreshes the Access Token.
   *
   * If a refresh is already in progress, this function will return immediately.
   * Once the refresh completes, all pending requests will resume.
   *
   * @async
   */
  async function refreshAccessToken() {
    if (isRefreshing) {
      return new Promise<void>((resolve, reject) => {
        pendingRequests.push((err?: Error) => {
          if (err) reject(err);
          else resolve();
        });
      });
    }

    isRefreshing = true;

    try {
      const res = await $baseFetch<ApiResponse<RefreshData>>("/auth/refresh", {
        method: "POST",
        credentials: "include",
      });

      // Update the token in Pinia store
      useAuthStore().setAccessToken(res.data.accessToken);

      pendingRequests.forEach((resolve) => resolve());
    } catch (err) {
      pendingRequests.forEach((resolve) => resolve(err as Error));
      throw err;
    } finally {
      isRefreshing = false;
      pendingRequests.splice(0);
    }
  }

  const get = <T>(url: string, options: FetchOptions<"json", unknown> = {}) => {
    return apiFetch<T>(url, { ...options, method: "get" });
  };

  const post = <Res>(url: string, body?: FetchBody, opts: FetchOptions<"json", unknown> = {}) =>
    apiFetch<Res>(url, { ...opts, method: "post", body });

  const put = <Res>(url: string, body?: FetchBody, opts: FetchOptions<"json", unknown> = {}) =>
    apiFetch<Res>(url, { ...opts, method: "put", body });

  const patch = <Res>(url: string, body?: FetchBody, opts: FetchOptions<"json", unknown> = {}) =>
    apiFetch<Res>(url, { ...opts, method: "patch", body });

  const del = <T>(url: string, options: FetchOptions<"json", unknown> = {}) => {
    return apiFetch<T>(url, { ...options, method: "delete" });
  };

  return { apiFetch, get, post, put, del, patch };
}
