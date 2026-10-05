export default defineNuxtRouteMiddleware(async (to) => {
  if (import.meta.server) {
    return;
  }

  if (!to.path.startsWith("/admin")) return;

  const auth = useAuthStore();

  if (auth.accessToken) return;

  try {
    await auth.refresh();
  } catch {
    auth.logout();

    return navigateTo("/auth/login");
  }
});
