import type { ApiResponse } from "~/types/api";
import type { User } from "~/types/user";

export const useAuthStore = defineStore("auth", () => {
  const user = ref<User | null>(null);
  const authChecked = ref(false);

  const isAuthenticated = computed(() => user.value !== null);

  let pendingFetch: Promise<User | null> | null = null;

  const fetchUser = async (force = false): Promise<User | null> => {
    if (!force && authChecked.value) return Promise.resolve(user.value);

    if (pendingFetch) {
      return pendingFetch;
    }

    const { $api } = useNuxtApp();

    pendingFetch = (async () => {
      try {
        const response = await $api<ApiResponse<User>>("/user/me");

        if (response.code !== 0) {
          throw new Error(`Failed to fetch user: ${response.code}`);
        }

        user.value = response.data;
        authChecked.value = true;

        return user.value;
      } catch (error) {
        const status = (error as { status?: number }).status;
        if (status === 401) {
          user.value = null;
          authChecked.value = true;
          return null;
        }
        throw error;
      }
    })().finally(() => {
      pendingFetch = null;
    });

    return pendingFetch;
  };

  const login = async (email: string, password: string) => {
    const { $api } = useNuxtApp();
    const response = await $api<ApiResponse<User>>("/auth/login", {
      method: "post",
      body: { email, password },
    });
    if (response.code !== 0) {
      throw new Error(`Login failed: ${response.code}`);
    }

    authChecked.value = false;

    const currentUser = await fetchUser(true);

    if (!currentUser) {
      throw new Error(`Failed to restore authenticated user.`);
    }

    await navigateTo("/admin");
  };

  const logout = async () => {
    const { $api } = useNuxtApp();
    try {
      await $api("/auth/logout", {
        method: "post",
      });
    } finally {
      user.value = null;
      authChecked.value = true;
    }

    await navigateTo("/auth");
  };

  const clear = () => {
    user.value = null;
    authChecked.value = false;
  };

  return {
    user,
    authChecked,
    isAuthenticated,
    fetchUser,
    login,
    logout,
    clear,
  };
});
