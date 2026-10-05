import type { ApiResponse } from "~/types/api";
import type { User } from "~/types/user";

export type AuthData = {
  accessToken: string;
} & User;

export const useAuthStore = defineStore("auth", () => {
  const { data } = useFetch<{ initialized: boolean }>("/api/v1/setup", {
    method: "GET",
  });

  const initialized = computed(() => data.value?.initialized ?? null);

  const { apiFetch } = useClientApi();

  const accessToken = ref(import.meta.client ? localStorage.getItem("accessToken") || "" : "");

  const setAccessToken = (newAccessToken: string) => {
    accessToken.value = newAccessToken;
    localStorage.setItem("accessToken", newAccessToken);
  };

  const logout = async () => {
    try {
      await apiFetch("/auth/logout", {
        method: "POST",
      });
    } finally {
      accessToken.value = "";
      localStorage.removeItem("accessToken");
    }
  };

  const login = async (email: string, password: string) => {
    const { code, data } = await apiFetch<ApiResponse<AuthData>>("/auth/login", {
      method: "POST",
      body: { email, password },
    });
    if (code !== 200) {
      // TODO: login failed logic
      console.warn(code);
    }
    setAccessToken(data.accessToken);

    // TODO: login successed logic

    navigateTo("/admin");
  };

  async function refresh() {
    const res = await $fetch<ApiResponse<{ accessToken: string }>>("/api/v1/auth/refresh", {
      method: "POST",
      credentials: "include",
    });

    accessToken.value = res.data.accessToken;
    localStorage.setItem("accessToken", res.data.accessToken);
  }

  return {
    initialized,
    accessToken,
    setAccessToken,
    refresh,
    login,
    logout,
  };
});
