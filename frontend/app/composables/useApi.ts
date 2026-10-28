import type { UseFetchOptions } from "#app";

type UseFetch = typeof useFetch;

const useApiFetchImpl = (
  url: string | Request | Ref<string | Request> | (() => string | Request),
  options?: UseFetchOptions<unknown>,
) => {
  const { $api } = useNuxtApp();

  return useFetch(url, {
    ...options,
    $fetch: $api,
  });
};

export const useApiFetch = useApiFetchImpl as UseFetch;
